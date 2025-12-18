package opencorpora

import "strings"

// SearchByLemmaAndPOS returns all entries that match the lemma (case-insensitive) and part of speech.
func SearchByLemmaAndPOS(entries []WordEntry, lemma string, pos PartOfSpeech) []WordEntry {
    normalizedLemma := strings.TrimSpace(lemma)
    if normalizedLemma == "" {
        return nil
    }
    out := make([]WordEntry, 0)
    for _, e := range entries {
        if e.POS == pos && strings.EqualFold(e.Lemma, normalizedLemma) {
            out = append(out, e)
        }
    }
    return out
}

// FilterByGrammemes returns entries that contain all specified grammemes.
func FilterByGrammemes(entries []WordEntry, grams ...Grammeme) []WordEntry {
    if len(grams) == 0 {
        return entries
    }
    wanted := make(map[Grammeme]struct{}, len(grams))
    for _, g := range grams {
        wanted[g] = struct{}{}
    }

    out := make([]WordEntry, 0)
    for _, e := range entries {
        if containsAll(e.Grammemes, wanted) {
            out = append(out, e)
        }
    }
    return out
}

// FilterStreamByGrammemes filters a stream of entries by grammemes.
// The returned channel is closed when input channel is exhausted.
func FilterStreamByGrammemes(entries <-chan WordEntry, grams ...Grammeme) <-chan WordEntry {
    out := make(chan WordEntry)
    wanted := make(map[Grammeme]struct{}, len(grams))
    for _, g := range grams {
        wanted[g] = struct{}{}
    }

    go func() {
        defer close(out)
        for e := range entries {
            if containsAll(e.Grammemes, wanted) {
                out <- e
            }
        }
    }()
    return out
}

func containsAll(have []Grammeme, wanted map[Grammeme]struct{}) bool {
    if len(wanted) == 0 {
        return true
    }
    haveSet := make(map[Grammeme]struct{}, len(have))
    for _, g := range have {
        haveSet[g] = struct{}{}
    }
    for g := range wanted {
        if _, ok := haveSet[g]; !ok {
            return false
        }
    }
    return true
}
