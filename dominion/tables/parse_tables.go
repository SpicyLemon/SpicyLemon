package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func PrintUsage() {
	fmt.Printf(`Usage: parse_tables.go [<source dir> [<dest dir>]

Parses the cleaned html tables into JSON.

All .html files in the <source dir> are read and parsed.
Default <source dir> is '.' (the current dir).
Output goes to <dest dir>.
Default <dest dir> is the <source dir>.
To provide a <dest dir> you must first provide a <source dir>.

Output is one .json file for each .html file plus one .json file containing everything.
`)
}

const notInTag = -1

type Card struct {
	Name        string   `json:"name"`
	Types       []string `json:"types"`
	Cost        string   `json:"cost"`
	Description string   `json:"description"`
}

type CardSet struct {
	Name  string  `json:"name"`
	Info  string  `json:"info"`
	Cards []*Card `json:"cards"`
}

var cleanRx = regexp.MustCompile(`[^a-zA-Z0-9.]+`)

func (s CardSet) Key() string {
	return cleanRx.ReplaceAllString(strings.ToLower(s.Name), "-")
}

func ConvertDir(sourceDir, destDir string) error {
	files, err := GetHtmlFiles(sourceDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no html files found in %q", sourceDir)
	}
	all := make(map[string]json.RawMessage)
	for _, filename := range files {
		contents, err := ParseFile(filename)
		if err != nil {
			return fmt.Errorf("error parsing %q: %w", filename, err)
		}
		game := GetFilenameBase(filename)
		bz, err := WriteFile(destDir, game, contents)
		if err != nil {
			return err
		}
		all[game] = bz
	}
	_, err = WriteFile(destDir, "all", all)
	return nil
}

func WriteFile(destDir, game string, contents map[string]json.RawMessage) ([]byte, error) {
	filename := filepath.Join(destDir, game+".json")
	bz, err := json.MarshalIndent(contents, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error creating JSON for %q: %w", filename, err)
	}
	err = os.WriteFile(filename, bz, 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing %q: %w", filename, err)
	}
	fmt.Printf("File created: %s\n", filename)
	return bz, nil
}

func ParseFile(filename string) (map[string]json.RawMessage, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")
	fmt.Printf("%s has %d lines.\n", filename, len(lines))
	tables, err := SplitTables(lines)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s has %d tables.\n", filename, len(tables))
	var cardSets []*CardSet
	for i, table := range tables {
		cardSet, err := ParseTable(table)
		if err != nil {
			return nil, fmt.Errorf("error parsing table %d: %w", i+1, err)
		}
		if cardSet.Name == "" {
			cardSet.Name = "Kingdom"
		}
		cardSets = append(cardSets, cardSet)
	}
	rv := make(map[string]json.RawMessage)
	for i, cardSet := range cardSets {
		bz, err := json.MarshalIndent(cardSet, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error creating JSON of card set %d: %w", i, err)
		}
		rv[cardSet.Key()] = bz
	}
	return rv, nil
}

func ParseTable(lines []string) (*CardSet, error) {
	if len(lines) < 2 {
		return nil, fmt.Errorf("unknown table contents: %q", strings.Join(lines, "\n"))
	}
	rv := &CardSet{}
	if strings.HasPrefix(lines[0], "<!--") && strings.HasSuffix(lines[0], "-->") {
		rv.Name = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(lines[0], "-->"), "<!--"))
		lines = lines[1:]
	}
	if strings.HasPrefix(lines[0], "<!--") && strings.HasSuffix(lines[0], "-->") {
		rv.Info = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[0], "<!--"), "-->"))
		lines = lines[1:]
	}
	rows, err := SplitRows(lines)
	if err != nil {
		return nil, err
	}
	for i, row := range rows {
		card, err := ParseRow(row)
		if err != nil {
			return nil, fmt.Errorf("error parsing row %d: %w", i+1, err)
		}
		rv.Cards = append(rv.Cards, card)
	}
	return rv, nil
}

func ParseRow(lines []string) (*Card, error) {
	if len(lines) < 4 {
		return nil, fmt.Errorf("not enough lines (%d) to be a card: %q", len(lines), strings.Join(lines, "\n"))
	}
	rv := &Card{}
	if !strings.HasPrefix(lines[0], "<td>") || !strings.HasSuffix(lines[0], "</td>") {
		return nil, fmt.Errorf("Unknown row line, expecting name: %q", lines[0])
	}
	rv.Name = strings.TrimSuffix(strings.TrimPrefix(lines[0], "<td>"), "</td>")
	lines = lines[1:]
	if !strings.HasPrefix(lines[0], "<td>") || !strings.HasSuffix(lines[0], "</td>") {
		return nil, fmt.Errorf("Unknown row line, expecting types: %q", lines[0])
	}
	rv.Types = strings.Split(strings.TrimSuffix(strings.TrimPrefix(lines[0], "<td>"), "</td>"), " -- ")
	lines = lines[1:]
	if !strings.HasPrefix(lines[0], "<td>") || !strings.HasSuffix(lines[0], "</td>") {
		return nil, fmt.Errorf("Unknown row line, expecting cost: %q", lines[0])
	}
	rv.Cost = strings.TrimSuffix(strings.TrimPrefix(lines[0], "<td>"), "</td>")
	lines = lines[1:]
	final := strings.Join(lines, "\n")
	if !strings.HasPrefix(final, "<td>") || !strings.HasSuffix(final, "</td>") {
		return nil, fmt.Errorf("Unknown row lines, expecting description: %q", final)
	}
	rv.Description = strings.TrimSuffix(strings.TrimPrefix(final, "<td>"), "</td>")
	return rv, nil
}

func SplitRows(lines []string) ([][]string, error) {
	var rv [][]string
	firstLine := notInTag
	for i, line := range lines {
		if line == "" {
			continue
		}
		if firstLine == notInTag {
			if line == "<tr>" {
				firstLine = i + 1
			} else {
				return nil, fmt.Errorf("Table line %d is %q but should be an opening row tag.", i, line)
			}
		} else {
			if line == "<tr>" {
				return nil, fmt.Errorf("Table line %d is %q but the previous row wasn't clsoed.", i, line)
			}
			if line == "</tr>" {
				rv = append(rv, lines[firstLine:i])
				firstLine = notInTag
			}
		}
	}
	if firstLine != notInTag {
		return nil, fmt.Errorf("Row open at the end of the table.")
	}
	return rv, nil
}

func SplitTables(lines []string) ([][]string, error) {
	var rv [][]string
	firstLine := notInTag
	for i, line := range lines {
		if line == "" {
			continue
		}
		if firstLine == notInTag {
			if line == "<table>" {
				firstLine = i + 1
			} else {
				return nil, fmt.Errorf("Line %d is %q but should be an opening table tag.", i, line)
			}
		} else {
			if line == "<table>" {
				return nil, fmt.Errorf("Line %d is %q but the previous table wasn't closed.", i, line)
			}
			if line == "</table>" {
				rv = append(rv, lines[firstLine:i])
				firstLine = notInTag
			}
		}
	}
	if firstLine != notInTag {
		return nil, fmt.Errorf("Table open at the end of the file.")
	}
	return rv, nil
}

func GetFilenameBase(filename string) string {
	base := filepath.Base(filename)
	parts := strings.Split(base, ".")
	return parts[0]
}

func GetHtmlFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read dir %q: %w", dir, err)
	}
	rv := make([]string, 0)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".html") {
			rv = append(rv, filepath.Join(dir, file.Name()))
		}
	}
	return rv, nil
}

func ParseArgs(args []string) (sourceDir string, destDir string, err error) {
	sourceDir = "."
	if len(args) > 0 {
		sourceDir = args[0]
	}
	destDir = sourceDir
	if len(args) > 1 {
		destDir = args[1]
	}
	if len(args) > 2 {
		err = fmt.Errorf("too many args, expected: 0 to 2, found: %d", len(args))
	}
	return
}

func main() {
	err := Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func Run() error {
	sourceDir, destDir, err := ParseArgs(os.Args[1:])
	if err != nil {
		PrintUsage()
		return err
	}

	return ConvertDir(sourceDir, destDir)
}
