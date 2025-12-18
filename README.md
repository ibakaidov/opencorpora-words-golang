# opencorpora words (Go)

Модуль для скачивания, парсинга и строготипизированной работы со словарём OpenCorpora.

## Особенности
- Загрузка и кэширование словаря (`~/.cache/opencorpora/dict.opcorpora.txt` по умолчанию).
- Парсинг в структуры `WordEntry` с полями: `LexemeID`, `Form`, `Lemma`, `POS` (enum), `Grammemes` (enum slice).
- Строгая типизация: части речи (`PartOfSpeech`) и граммемы (`Grammeme`) генерируются из исходного словаря (`go generate`).
- Поиск по начальной форме и части речи, фильтрация по граммемам, потоковая обработка через каналы.
- Тесты на примере упрощённого словаря в `testdata/`.

## Установка
```bash
go get ibakaidov/opencorpora-words-golang
```

## Генерация enum-ов
Парсер опирается на реальные значения из словаря, поэтому перед использованием в CI/локально выполните:
```bash
cd opencorpora
go generate
```
`go:generate` скачает/обновит архив, распакует его в кэш и сгенерирует `opencorpora/enums_gen.go`.

## Использование
```go
ctx := context.Background()

// 1. Убедиться, что словарь скачан и распакован.
dictPath, err := opencorpora.EnsureDictionary(ctx)
if err != nil {
    log.Fatal(err)
}

// 2a. Загрузить всё в память.
entries, err := opencorpora.ParseFile(ctx, dictPath)
if err != nil {
    log.Fatal(err)
}

// 2b. Или потоково обрабатывать (меньше памяти).
stream, errs := opencorpora.StreamFile(ctx, dictPath)
for e := range opencorpora.FilterStreamByGrammemes(stream, opencorpora.GrammemePlur) {
    fmt.Println(e.Form, e.Grammemes)
}
if err := <-errs; err != nil {
    log.Fatal(err)
}

// 3. Поиск и фильтрация.
forms := opencorpora.SearchByLemmaAndPOS(entries, "кот", opencorpora.PartNoun)
plurGent := opencorpora.FilterByGrammemes(entries, opencorpora.GrammemePlur, opencorpora.GrammemeGent)
```

### CLI-пример
```bash
# вывести доступные части речи и граммемы
go run ./cmd/opencorpora-cli -list-pos
go run ./cmd/opencorpora-cli -list-grammemes

# найти формы леммы "кот" (NOUN) в именительном падеже
go run ./cmd/opencorpora-cli -lemma кот -pos NOUN -grammemes nomn -limit 10

# потоковая фильтрация всех множественных генитивных форм (меньше памяти)
go run ./cmd/opencorpora-cli -stream -grammemes plur,gent -limit 20
```

## Пользовательские настройки
- `WithCacheDir(dir string)` — задать другой кэш.
- `WithDictionaryURL(url string)` — альтернативный источник словаря.
- `WithZipPath/WithTextPath` — использовать уже скачанные файлы.

## Тесты
```bash
go test ./...
```

## Структура файла словаря
Секция начинается с идентификатора леммы (целое число), далее идут строки вида `СЛОВО<TAB>POS,граммемы`. Пустые строки разделяют записи. Начальная форма (`Lemma`) определяется как первое слово в секции.
