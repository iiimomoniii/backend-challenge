package code

import "testing"

func TestLookup(t *testing.T) {
	t.Run("known code returns its entry", func(t *testing.T) {
		entry, ok := Lookup("USR006")
		if !ok {
			t.Fatal("want ok=true for known code USR006")
		}
		if entry.Name != "USER_NOT_FOUND" {
			t.Errorf("want Name %q, got %q", "USER_NOT_FOUND", entry.Name)
		}
		if entry.MessageEN == "" {
			t.Error("want non-empty MessageEN")
		}
		if entry.MessageTH == "" {
			t.Error("want non-empty MessageTH")
		}
	})

	t.Run("unknown code returns ok=false", func(t *testing.T) {
		_, ok := Lookup("USR999")
		if ok {
			t.Fatal("want ok=false for unknown code USR999")
		}
	})
}

func TestMessage(t *testing.T) {
	t.Run("known code in English", func(t *testing.T) {
		got := Message("USR006", LangEN)
		want := "User not found"
		if got != want {
			t.Errorf("want %q, got %q", want, got)
		}
	})

	t.Run("known code in Thai", func(t *testing.T) {
		got := Message("USR006", LangTH)
		want := "ไม่พบผู้ใช้งาน"
		if got != want {
			t.Errorf("want %q, got %q", want, got)
		}
	})

	t.Run("unknown code falls back to the raw code string", func(t *testing.T) {
		got := Message("USR999", LangEN)
		if got != "USR999" {
			t.Errorf("want the raw code %q back, got %q", "USR999", got)
		}
	})

	t.Run("every catalog entry has both languages populated", func(t *testing.T) {
		for id, entry := range catalog {
			if entry.MessageEN == "" {
				t.Errorf("catalog entry %s is missing MessageEN", id)
			}
			if entry.MessageTH == "" {
				t.Errorf("catalog entry %s is missing MessageTH", id)
			}
		}
	})
}

func TestParseLang(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Lang
	}{
		{"lowercase th", "th", LangTH},
		{"uppercase TH", "TH", LangTH},
		{"th-TH locale tag", "th-TH", LangTH},
		{"empty string defaults to English", "", LangEN},
		{"unrecognized value defaults to English", "fr", LangEN},
		{"lowercase en", "en", LangEN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLang(tt.raw)
			if got != tt.want {
				t.Errorf("ParseLang(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
