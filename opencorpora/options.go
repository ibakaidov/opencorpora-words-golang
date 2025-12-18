package opencorpora

import "context"

const (
    // DefaultDictionaryURL points to the public OpenCorpora export.
    DefaultDictionaryURL = "https://opencorpora.org/files/export/dict/dict.opcorpora.txt.zip"
    cacheDirName         = "opencorpora"
    zipFileName          = "dict.opcorpora.txt.zip"
    textFileName         = "dict.opcorpora.txt"
)

// WithCacheDir overrides the cache directory where the zip and extracted files will be stored.
func WithCacheDir(dir string) Option {
    return func(o *Options) { o.CacheDir = dir }
}

// WithDictionaryURL overrides the URL used to download the dictionary archive.
func WithDictionaryURL(url string) Option {
    return func(o *Options) { o.DictionaryURL = url }
}

// WithZipPath forces the loader to use the given zip archive path instead of the default cache path.
func WithZipPath(path string) Option {
    return func(o *Options) { o.ZipPath = path }
}

// WithTextPath forces the loader to use the given extracted dictionary file instead of the default cache path.
func WithTextPath(path string) Option {
    return func(o *Options) { o.TextPath = path }
}

// newOptions builds an Options value applying provided Option functions and defaults.
func newOptions(ctx context.Context, opts []Option) (Options, error) {
    o := Options{}
    for _, fn := range opts {
        fn(&o)
    }
    if err := o.Apply(ctx); err != nil {
        return Options{}, err
    }
    return o, nil
}
