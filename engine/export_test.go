package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExportAllTranslationsMultiline(t *testing.T) {
	tests := []struct {
		name string
		data KeyLangMap
		want string
	}{
		{
			name: "singular",
			data: KeyLangMap{"Hello": {"en": {OneValue: "Hello there"}}},
			want: "{\n\t\"Hello\": \"Hello there\"\n}\n",
		},
		{
			name: "plural",
			data: KeyLangMap{"Items": {"en": {OneValue: "One item", OtherValue: "Many items", IsPlural: true}}},
			want: "{\n\t\"Items\": {\n\t\t\"One\": \"One item\",\n\t\t\"Other\": \"Many items\"\n\t}\n}\n",
		},
		{
			name: "escaped strings",
			data: KeyLangMap{"Say \"hello\"": {"en": {OneValue: "Line one\nLine two\tC:\\temp"}}},
			want: "{\n\t\"Say \\\"hello\\\"\": \"Line one\\nLine two\\tC:\\\\temp\"\n}\n",
		},
		{
			name: "empty",
			data: KeyLangMap{},
			want: "{}\n",
		},
		{
			name: "skip untranslated",
			data: KeyLangMap{"Missing": {"en": {OneValue: ""}}},
			want: "{}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project := ProjectNew(t.TempDir(), ".", []string{"en"})
			if err := ExportAllTranslations(&project, test.data); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(filepath.Join(project.LocationPath, "en.json"))
			if err != nil {
				t.Fatal(err)
			}
			if string(content) != test.want {
				t.Fatalf("export = %q, want %q", content, test.want)
			}
			if !json.Valid(content) {
				t.Fatalf("invalid JSON: %s", content)
			}
		})
	}
}

func TestExportAllTranslationsMultipleKeysAndLanguages(t *testing.T) {
	project := ProjectNew(t.TempDir(), ".", []string{"en", "fr"})
	data := KeyLangMap{
		"Greeting": {
			"en": {OneValue: "Hello"},
			"fr": {OneValue: "Bonjour"},
		},
		"Items": {
			"en": {OneValue: "One \"item\"", OtherValue: "Many\nitems", IsPlural: true},
			"fr": {OneValue: "Un objet", OtherValue: "Des objets", IsPlural: true},
		},
	}
	if err := ExportAllTranslations(&project, data); err != nil {
		t.Fatal(err)
	}
	for _, lang := range project.ManagedLanguages {
		content, err := os.ReadFile(filepath.Join(project.LocationPath, string(lang)+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), ",\n\t\"") || !strings.HasSuffix(string(content), "\n}\n") {
			t.Fatalf("expected multiline export for %s: %s", lang, content)
		}
		var got map[string]any
		if err := json.Unmarshal(content, &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{
			"Greeting": data["Greeting"][lang].OneValue,
			"Items": map[string]any{
				"One": data["Items"][lang].OneValue, "Other": data["Items"][lang].OtherValue,
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s export = %#v, want %#v", lang, got, want)
		}
	}
}
