package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"ibakaidov/opencorpora-words-golang/opencorpora"
)

func main() {
	var (
		lemma     = flag.String("lemma", "", "Начальная форма для поиска (необязательно)")
		posStr    = flag.String("pos", "", "Часть речи (например, NOUN). Можно узнать через -list-pos")
		gramsStr  = flag.String("grammemes", "", "Список граммем через запятую (например, plur,gent)")
		useStream = flag.Bool("stream", false, "Стриминговый парсинг (меньше памяти, возможно медленнее)")
		limit     = flag.Int("limit", 0, "Ограничить число выводимых строк (0 — без ограничений)")

		cacheDir = flag.String("cache", "", "Кэш для словаря (по умолчанию ~/.cache/opencorpora)")
		dictURL  = flag.String("dict-url", opencorpora.DefaultDictionaryURL, "URL словаря")
		zipPath  = flag.String("zip", "", "Путь к готовому zip (опционально)")
		textPath = flag.String("text", "", "Путь к распакованному словарю (опционально)")

		listPOS  = flag.Bool("list-pos", false, "Вывести доступные части речи и выйти")
		listGram = flag.Bool("list-grammemes", false, "Вывести доступные граммемы и выйти")
	)

	flag.Parse()

	if *listPOS {
		for _, p := range opencorpora.AllPartsOfSpeech() {
			fmt.Println(p.String())
		}
		return
	}
	if *listGram {
		for _, g := range opencorpora.AllGrammemes() {
			fmt.Println(g.String())
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := []opencorpora.Option{opencorpora.WithDictionaryURL(*dictURL)}
	if *cacheDir != "" {
		opts = append(opts, opencorpora.WithCacheDir(*cacheDir))
	}
	if *zipPath != "" {
		opts = append(opts, opencorpora.WithZipPath(*zipPath))
	}
	if *textPath != "" {
		opts = append(opts, opencorpora.WithTextPath(*textPath))
	}

	dictPath, err := opencorpora.EnsureDictionary(ctx, opts...)
	if err != nil {
		log.Fatalf("ensure dictionary: %v", err)
	}

	var pos opencorpora.PartOfSpeech
	if *posStr != "" {
		var ok bool
		pos, ok = opencorpora.ParsePartOfSpeech(strings.TrimSpace(*posStr))
		if !ok {
			log.Fatalf("unknown part of speech: %s", *posStr)
		}
	}

	grams, err := parseGrammemes(*gramsStr)
	if err != nil {
		log.Fatal(err)
	}

	if *useStream {
		runStream(ctx, dictPath, *lemma, pos, *posStr != "", grams, *limit)
	} else {
		runInMemory(ctx, dictPath, *lemma, pos, *posStr != "", grams, *limit)
	}
}

func runInMemory(ctx context.Context, dictPath, lemma string, pos opencorpora.PartOfSpeech, hasPOS bool, grams []opencorpora.Grammeme, limit int) {
	entries, err := opencorpora.ParseFile(ctx, dictPath)
	if err != nil {
		log.Fatalf("parse dictionary: %v", err)
	}

	filtered := entries
	if lemma != "" && hasPOS {
		filtered = opencorpora.SearchByLemmaAndPOS(filtered, lemma, pos)
	}
	if len(grams) > 0 {
		filtered = opencorpora.FilterByGrammemes(filtered, grams...)
	}
	if lemma != "" && !hasPOS {
		filtered = filterLemma(filtered, lemma)
	}
	if hasPOS && lemma == "" {
		filtered = filterPOS(filtered, pos)
	}

	emitEntries(filtered, limit)
}

func runStream(ctx context.Context, dictPath, lemma string, pos opencorpora.PartOfSpeech, hasPOS bool, grams []opencorpora.Grammeme, limit int) {
	stream, errs := opencorpora.StreamFile(ctx, dictPath)
	filtered := stream
	if len(grams) > 0 {
		filtered = opencorpora.FilterStreamByGrammemes(filtered, grams...)
	}

	count := 0
	for e := range filtered {
		if lemma != "" && !strings.EqualFold(e.Lemma, lemma) {
			continue
		}
		if hasPOS && e.POS != pos {
			continue
		}
		fmt.Println(formatEntry(e))
		count++
		if limit > 0 && count >= limit {
			break
		}
	}
	if err := <-errs; err != nil {
		log.Fatalf("stream: %v", err)
	}
}

func parseGrammemes(s string) ([]opencorpora.Grammeme, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	grams := make([]opencorpora.Grammeme, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		g, ok := opencorpora.ParseGrammeme(token)
		if !ok {
			return nil, fmt.Errorf("unknown grammeme: %s", token)
		}
		grams = append(grams, g)
	}
	return grams, nil
}

func emitEntries(entries []opencorpora.WordEntry, limit int) {
	count := 0
	for _, e := range entries {
		fmt.Println(formatEntry(e))
		count++
		if limit > 0 && count >= limit {
			break
		}
	}
}

func formatEntry(e opencorpora.WordEntry) string {
	grams := make([]string, len(e.Grammemes))
	for i, g := range e.Grammemes {
		grams[i] = g.String()
	}
	return fmt.Sprintf("%d\t%s\t%s\t%s\t%s", e.LexemeID, e.Lemma, e.Form, e.POS.String(), strings.Join(grams, ","))
}

func filterLemma(entries []opencorpora.WordEntry, lemma string) []opencorpora.WordEntry {
	out := make([]opencorpora.WordEntry, 0)
	for _, e := range entries {
		if strings.EqualFold(e.Lemma, lemma) {
			out = append(out, e)
		}
	}
	return out
}

func filterPOS(entries []opencorpora.WordEntry, pos opencorpora.PartOfSpeech) []opencorpora.WordEntry {
	out := make([]opencorpora.WordEntry, 0)
	for _, e := range entries {
		if e.POS == pos {
			out = append(out, e)
		}
	}
	return out
}
