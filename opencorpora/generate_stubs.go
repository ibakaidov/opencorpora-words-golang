//go:build generate
// +build generate

package opencorpora

// The functions below allow `go run` during code generation before enums are produced.

func ParsePartOfSpeech(string) (PartOfSpeech, bool) { return PartOfSpeech(0), false }

func ParseGrammeme(string) (Grammeme, bool) { return Grammeme(0), false }

func AllPartsOfSpeech() []PartOfSpeech { return nil }

func AllGrammemes() []Grammeme { return nil }
