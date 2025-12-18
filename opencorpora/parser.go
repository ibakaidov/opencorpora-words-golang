package opencorpora

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"
)

// ParseFile loads the dictionary file at the given path into memory.
func ParseFile(ctx context.Context, path string) ([]WordEntry, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open dictionary: %w", err)
    }
    defer f.Close()
    return Parse(ctx, f)
}

// Parse reads all entries from the given reader and returns them as a slice.
func Parse(ctx context.Context, r io.Reader) ([]WordEntry, error) {
    entries := make([]WordEntry, 0, 1024)
    out, errs := Stream(ctx, r)
    for e := range out {
        entries = append(entries, e)
    }
    if err := <-errs; err != nil {
        return nil, err
    }
    return entries, nil
}

// StreamFile streams entries from the file at the given path through a channel.
func StreamFile(ctx context.Context, path string) (<-chan WordEntry, <-chan error) {
    ch := make(chan WordEntry)
    errs := make(chan error, 1)

    go func() {
        defer close(ch)
        defer close(errs)

        f, err := os.Open(path)
        if err != nil {
            errs <- fmt.Errorf("open dictionary: %w", err)
            return
        }
        defer f.Close()

        out, innerErrs := Stream(ctx, f)
        for e := range out {
            ch <- e
        }
        if err := <-innerErrs; err != nil {
            errs <- err
        }
    }()

    return ch, errs
}

// Stream parses entries from the reader and sends them into a channel. The caller must read until the channel closes.
func Stream(ctx context.Context, r io.Reader) (<-chan WordEntry, <-chan error) {
    ch := make(chan WordEntry)
    errs := make(chan error, 1)

    go func() {
        defer close(ch)
        defer close(errs)

        scanner := bufio.NewScanner(r)
        // Lines are short, but keep some headroom.
        scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

        var lexemeID int
        var lemma string

        for scanner.Scan() {
            if err := ctx.Err(); err != nil {
                errs <- err
                return
            }
            line := strings.TrimSpace(scanner.Text())
            if line == "" {
                continue
            }

            if id, err := strconv.Atoi(line); err == nil {
                lexemeID = id
                lemma = ""
                continue
            }

            if lexemeID == 0 {
                errs <- fmt.Errorf("encountered form before lexeme id: %s", line)
                return
            }

            entry, err := parseLine(line)
            if err != nil {
                errs <- fmt.Errorf("lexeme %d: %w", lexemeID, err)
                return
            }

            if lemma == "" {
                lemma = entry.Form
            }
            entry.Lemma = lemma
            entry.LexemeID = lexemeID
            ch <- entry
        }

        if err := scanner.Err(); err != nil {
            errs <- fmt.Errorf("scan dictionary: %w", err)
            return
        }
        errs <- nil
    }()

    return ch, errs
}

func parseLine(line string) (WordEntry, error) {
    parts := strings.SplitN(line, "\t", 2)
    if len(parts) != 2 {
        return WordEntry{}, fmt.Errorf("malformed line (expected word\\tPOS,grammemes): %q", line)
    }

    form := strings.TrimSpace(parts[0])
    tokenStr := strings.TrimSpace(parts[1])
    if form == "" || tokenStr == "" {
        return WordEntry{}, fmt.Errorf("empty form or tags in line: %q", line)
    }

    tokens := splitTokens(tokenStr)
    if len(tokens) == 0 {
        return WordEntry{}, fmt.Errorf("missing tags in line: %q", line)
    }

    posToken := tokens[0]
    pos, ok := ParsePartOfSpeech(posToken)
    if !ok {
        return WordEntry{}, fmt.Errorf("unknown part of speech: %s", posToken)
    }

    grammemes := make([]Grammeme, 0, len(tokens)-1)
    seen := make(map[Grammeme]struct{})
    for _, g := range tokens[1:] {
        gram, ok := ParseGrammeme(g)
        if !ok {
            return WordEntry{}, fmt.Errorf("unknown grammeme: %s", g)
        }
        if _, exists := seen[gram]; !exists {
            grammemes = append(grammemes, gram)
            seen[gram] = struct{}{}
        }
    }

    return WordEntry{Form: form, POS: pos, Grammemes: grammemes}, nil
}

func splitTokens(s string) []string {
    s = strings.ReplaceAll(s, ",", " ")
    fields := strings.Fields(s)
    return fields
}
