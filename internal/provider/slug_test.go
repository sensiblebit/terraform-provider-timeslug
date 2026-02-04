package provider

import (
	"fmt"
	"testing"
)

// Reference test vectors - must match all implementations (Python, Java, C++)
var testVectors = []struct {
	seed, period, mode string
	length             int
	slug, hash         string
}{
	{"seedphrase", "2026-02-03", "obfuscated", 16, "trybeambold8", "5d3bf0d55db67ea2"},
	{"seedphrase", "2026-02-04", "obfuscated", 16, "brightbeamvivar", "f9fb66a05050f52f"},
	{"seedphrase", "2026-02-05", "obfuscated", 16, "trycorefastfum", "8bb68bd056e4a6ff"},
	{"seedphrase", "2026-02-03", "bip39", 3, "exoticangryanswer", "50011c26d0"},
	{"seedphrase", "2026-02-03", "bip39", 5, "exoticangryanswerpatternmain", "50011c26d0a864"},
}

func TestDerive(t *testing.T) {
	for _, tc := range testVectors {
		slug, hash := derive(tc.seed, tc.period, tc.length, tc.mode)
		if slug != tc.slug || hash != tc.hash {
			t.Errorf("%s/%s: got %q/%q, want %q/%q", tc.period, tc.mode, slug, hash, tc.slug, tc.hash)
		}
	}

	// Mode is case-insensitive
	s1, _ := derive("seed", "2026-01-01", 3, "bip39")
	s2, _ := derive("seed", "2026-01-01", 3, "BIP39")
	if s1 != s2 {
		t.Error("mode should be case insensitive")
	}
}

func TestGenerate(t *testing.T) {
	slugs, err := Generate("seedphrase", "2026-02-03", 16, 3, "day", "obfuscated")
	if err != nil {
		t.Fatal(err)
	}
	if len(slugs) != 3 || slugs[1].Value != "trybeambold8" {
		t.Errorf("got %d slugs, center=%q", len(slugs), slugs[1].Value)
	}
}

func TestGenerateErrors(t *testing.T) {
	if _, err := Generate("seed", "invalid", 3, 3, "day", "bip39"); err == nil {
		t.Error("expected error for invalid time")
	}
	if _, err := Generate("seed", "2026-02-03", 3, 3, "invalid", "bip39"); err == nil {
		t.Error("expected error for invalid interval")
	}
}

func TestParseTime(t *testing.T) {
	valid := []string{"2026-02-03", "2026-02-03T15", "2026-02-03T15:04", "2026-02-03T15:04:05", "2026-02-03T15:04:05Z"}
	for _, s := range valid {
		if _, err := parseTime(s); err != nil {
			t.Errorf("parseTime(%q) failed: %v", s, err)
		}
	}
	invalid := []string{"invalid", "02-03-2026", ""}
	for _, s := range invalid {
		if _, err := parseTime(s); err == nil {
			t.Errorf("parseTime(%q) should fail", s)
		}
	}
}

func TestParseInterval(t *testing.T) {
	valid := []string{"s", "second", "m", "minute", "h", "hour", "d", "day", "", "w", "week"}
	for _, s := range valid {
		if _, _, err := parseInterval(s); err != nil {
			t.Errorf("parseInterval(%q) failed: %v", s, err)
		}
	}
	if _, _, err := parseInterval("invalid"); err == nil {
		t.Error("parseInterval(invalid) should fail")
	}
}

func TestShortenWord(t *testing.T) {
	cases := map[string]string{
		"the": "the", "cat": "cat", // short unchanged
		"tiger": "tigr", "silver": "silvr", // -er pattern
		"purple": "purpl", "simple": "simpl", // -le pattern
		"delta": "delt", "alpha": "alph", // trailing vowel
		"nova": "nova", "cloud": "cloud", // no change
	}
	for in, want := range cases {
		if got := shortenWord(in); got != want {
			t.Errorf("shortenWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestContainsBlockedWord(t *testing.T) {
	blocked := []string{"shitty", "SHITTY", "hacker", "hello"} // hello contains "hell"
	for _, s := range blocked {
		if !containsBlockedWord(s) {
			t.Errorf("containsBlockedWord(%q) should be true", s)
		}
	}
	clean := []string{"greetings", "world", "testslug", "", "a"}
	for _, s := range clean {
		if containsBlockedWord(s) {
			t.Errorf("containsBlockedWord(%q) should be false", s)
		}
	}
}

func TestBuildObfuscatedSlug(t *testing.T) {
	// Known value
	entropy := hmacSHA256("seedphrase", "seedphrase:2026-02-03")
	if slug := buildObfuscatedSlug(entropy); slug != "trybeambold8" {
		t.Errorf("got %q, want trybeambold8", slug)
	}

	// Length constraints and invariants (10-18 chars, no blocked, no triple letters)
	for i := range 100 {
		entropy := hmacSHA256("test", fmt.Sprintf("%d", i))
		slug := buildObfuscatedSlug(entropy)
		if len(slug) < 10 || len(slug) > 18 {
			t.Errorf("slug %q length %d out of range", slug, len(slug))
		}
		if containsBlockedWord(slug) {
			t.Errorf("slug %q contains blocked word", slug)
		}
		for j := 2; j < len(slug); j++ {
			if slug[j] == slug[j-1] && slug[j] == slug[j-2] {
				t.Errorf("slug %q has triple letter", slug)
			}
		}
	}
}

func TestEntropyToBIP39Words(t *testing.T) {
	words := entropyToBIP39Words(hmacSHA256("seed", "seed:2026-01-01"))
	if len(words) != 24 {
		t.Errorf("expected 24 words, got %d", len(words))
	}
}

func TestBIP39WordlistLoaded(t *testing.T) {
	if len(bip39Words) != 2048 {
		t.Errorf("expected 2048 BIP39 words, got %d", len(bip39Words))
	}
}

func TestSlugUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 1; i <= 28; i++ {
		slug, _ := derive("seed", fmt.Sprintf("2026-01-%02d", i), 3, "bip39")
		if seen[slug] {
			t.Errorf("collision for slug %q", slug)
		}
		seen[slug] = true
	}
}

func BenchmarkDerive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		derive("seedphrase", "2026-02-03", 3, "bip39")
	}
}
