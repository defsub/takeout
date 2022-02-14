package pls

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/defsub/takeout/lib/str"
)

const (
	MaxEntries = 100
	DefaultVersion = 2
)

var (
	ErrMaxEntries    = errors.New("max entries exceeded")
	ErrInvalidFormat = errors.New("invalid format")
)

type Entry struct {
	Index  int
	File   string
	Title  string
	Length int
}

type Playlist struct {
	Version         int
	NumberOfEntries int
	Entries         []Entry
}

func parse(in string) (Playlist, error) {
	return Parse(strings.NewReader(in))
}

func Parse(in io.Reader) (Playlist, error) {
	scanner := bufio.NewScanner(in)

	// https://en.wikipedia.org/wiki/PLS_(file_format)
	entryRegexp := regexp.MustCompile(`(?i)(File|Title|Length)([\d]+)=(.+)`)
	versionRegexp := regexp.MustCompile(`(?i)Version=([\d]+)`)
	numberRegexp := regexp.MustCompile(`(?i)NumberOfEntries=([\d]+)`)

	entries := make(map[int]*Entry)
	numberOfEntries := 0
	version := DefaultVersion

	for scanner.Scan() {
		line := scanner.Text()
		matches := entryRegexp.FindStringSubmatch(line)
		if matches != nil {
			field := matches[1]
			index := str.Atoi(matches[2])
			value := matches[3]

			v, ok := entries[index]
			if !ok {
				v = &Entry{}
				v.Index = index
				entries[index] = v
			}

			switch f := strings.ToLower(field); f {
			case "file":
				v.File = value
			case "title":
				v.Title = value
			case "length":
				v.Length = str.Atoi(value)
			}
		}

		matches = versionRegexp.FindStringSubmatch(line)
		if matches != nil {
			version = str.Atoi(matches[1])
		}

		matches = numberRegexp.FindStringSubmatch(line)
		if matches != nil {
			numberOfEntries = str.Atoi(matches[1])
		}
	}

	var result []Entry

	if version != DefaultVersion ||
		numberOfEntries != len(entries) {
		return Playlist{}, ErrInvalidFormat
	}

	if numberOfEntries > MaxEntries {
		return Playlist{}, ErrMaxEntries
	}

	for i := 1; i <= numberOfEntries; i++ {
		entry, ok := entries[i]
		if !ok {
			break
		}
		result = append(result, *entry)
	}

	return Playlist{
		Version:         version,
		NumberOfEntries: numberOfEntries,
		Entries:         result}, nil
}
