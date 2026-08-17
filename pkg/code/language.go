package code

import (
	_ "embed"
	"encoding/json"
)

var errorCatalogRaw []byte

type Entry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MessageEN string `json:"messageEN"`
	MessageTH string `json:"messageTH"`
}

type Lang string

const (
	LangEN Lang = "en"
	LangTH Lang = "th"
)

var catalog map[string]Entry

func init() {
	if err := json.Unmarshal(errorCatalogRaw, &catalog); err != nil {
		panic("pkg/code: failed to parse error.json: " + err.Error())
	}
}

func Lookup(codeID string) (Entry, bool) {
	entry, ok := catalog[codeID]
	return entry, ok
}

func Message(codeID string, lang Lang) string {
	entry, ok := catalog[codeID]
	if !ok {
		return codeID
	}
	if lang == LangTH {
		return entry.MessageTH
	}
	return entry.MessageEN
}

func ParseLang(raw string) Lang {
	switch raw {
	case "th", "TH", "th-TH":
		return LangTH
	default:
		return LangEN
	}
}