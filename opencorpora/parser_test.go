package opencorpora

import (
    "context"
    "os"
    "testing"
)

func loadSample(t *testing.T) []WordEntry {
    t.Helper()
    f, err := os.Open("../testdata/sample_dict.txt")
    if err != nil {
        t.Fatalf("open sample: %v", err)
    }
    defer f.Close()

    entries, err := Parse(context.Background(), f)
    if err != nil {
        t.Fatalf("parse sample: %v", err)
    }
    return entries
}

func TestParseSample(t *testing.T) {
    entries := loadSample(t)
    if len(entries) != 6 {
        t.Fatalf("unexpected entries count: %d", len(entries))
    }

    first := entries[0]
    if first.LexemeID != 1 || first.Lemma != "КОТ" || first.Form != "КОТ" {
        t.Fatalf("unexpected first entry: %+v", first)
    }
    if first.POS != PartNoun {
        t.Fatalf("unexpected POS: %v", first.POS)
    }
    if !containsAll(first.Grammemes, map[Grammeme]struct{}{GrammemeAnim: {}, GrammemeMasc: {}, GrammemeSing: {}, GrammemeNomn: {}}) {
        t.Fatalf("unexpected grammemes: %+v", first.Grammemes)
    }

    last := entries[len(entries)-1]
    if last.LexemeID != 2 || last.Lemma != "БЫСТРЫЙ" {
        t.Fatalf("unexpected last entry: %+v", last)
    }
}

func TestSearchByLemmaAndPOS(t *testing.T) {
    entries := loadSample(t)

    res := SearchByLemmaAndPOS(entries, "кот", PartNoun)
    if len(res) != 3 {
        t.Fatalf("expected 3 noun forms for lemma КОТ, got %d", len(res))
    }

    resAdj := SearchByLemmaAndPOS(entries, "быстрый", PartAdjf)
    if len(resAdj) != 3 {
        t.Fatalf("expected 3 adjective forms for lemma БЫСТРЫЙ, got %d", len(resAdj))
    }
}

func TestFilterByGrammemes(t *testing.T) {
    entries := loadSample(t)

    res := FilterByGrammemes(entries, GrammemePlur, GrammemeGent)
    if len(res) != 1 {
        t.Fatalf("expected 1 entry with plural+genitive, got %d", len(res))
    }
    if res[0].Form != "БЫСТРЫХ" {
        t.Fatalf("unexpected form: %s", res[0].Form)
    }
}
