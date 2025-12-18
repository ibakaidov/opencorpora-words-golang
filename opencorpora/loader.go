package opencorpora

import (
    "archive/zip"
    "context"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

func defaultCacheDir() (string, error) {
    base, err := os.UserCacheDir()
    if err != nil {
        return "", fmt.Errorf("cannot resolve user cache dir: %w", err)
    }
    return filepath.Join(base, cacheDirName), nil
}

func defaultZipPath(cacheDir string) string {
    return filepath.Join(cacheDir, zipFileName)
}

func defaultTextPath(cacheDir string) string {
    return filepath.Join(cacheDir, textFileName)
}

// EnsureDictionary downloads (if missing) and extracts the dictionary, returning the path to the text file.
func EnsureDictionary(ctx context.Context, opts ...Option) (string, error) {
    o, err := newOptions(ctx, opts)
    if err != nil {
        return "", err
    }

    if err := os.MkdirAll(o.CacheDir, 0o755); err != nil {
        return "", fmt.Errorf("create cache dir: %w", err)
    }

    if _, err := os.Stat(o.TextPath); err == nil {
        return o.TextPath, nil
    }

    if _, err := os.Stat(o.ZipPath); errors.Is(err, os.ErrNotExist) {
        if err := download(ctx, o.DictionaryURL, o.ZipPath); err != nil {
            return "", err
        }
    }

    if err := extractZip(ctx, o.ZipPath, o.TextPath); err != nil {
        return "", err
    }
    return o.TextPath, nil
}

func download(ctx context.Context, url, dest string) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return fmt.Errorf("build request: %w", err)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("download dictionary: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("download dictionary: unexpected status %s", resp.Status)
    }

    tmp := dest + ".tmp"
    if err := func() error {
        f, err := os.Create(tmp)
        if err != nil {
            return fmt.Errorf("create temp file: %w", err)
        }
        defer f.Close()
        if _, err := io.Copy(f, resp.Body); err != nil {
            return fmt.Errorf("write temp file: %w", err)
        }
        return nil
    }(); err != nil {
        return err
    }

    if err := os.Rename(tmp, dest); err != nil {
        return fmt.Errorf("move temp file: %w", err)
    }
    return nil
}

func extractZip(ctx context.Context, zipPath, dest string) error {
    zf, err := zip.OpenReader(zipPath)
    if err != nil {
        return fmt.Errorf("open zip: %w", err)
    }
    defer zf.Close()

    if len(zf.File) == 0 {
        return fmt.Errorf("zip archive %s is empty", zipPath)
    }

    // Use the first file in the archive.
    f := zf.File[0]
    rc, err := f.Open()
    if err != nil {
        return fmt.Errorf("open zip entry: %w", err)
    }
    defer rc.Close()

    tmp := dest + ".tmp"
    if err := func() error {
        out, err := os.Create(tmp)
        if err != nil {
            return fmt.Errorf("create extracted file: %w", err)
        }
        defer out.Close()
        if _, err := io.Copy(out, rc); err != nil {
            return fmt.Errorf("write extracted file: %w", err)
        }
        return nil
    }(); err != nil {
        return err
    }

    if err := ctx.Err(); err != nil {
        return err
    }

    if err := os.Rename(tmp, dest); err != nil {
        return fmt.Errorf("move extracted file: %w", err)
    }
    return nil
}
