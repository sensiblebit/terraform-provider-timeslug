package provider

import (
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
)

//go:embed bip39_english.txt
var bip39Raw string
var bip39Words []string

func init() {
	bip39Words = strings.Split(strings.TrimSpace(bip39Raw), "\n")
}

type Slug struct {
	Value  string
	Period string
	Hash   string
}

// Generate creates slugs for a time window centered on anchor.
func Generate(seed, anchor string, length, window int, interval, mode string) ([]Slug, error) {
	anchorTime, err := parseTime(anchor)
	if err != nil {
		return nil, err
	}
	duration, format, err := parseInterval(interval)
	if err != nil {
		return nil, err
	}

	slugs := make([]Slug, window)
	start := -window / 2
	for i := range slugs {
		period := anchorTime.Add(time.Duration(start+i) * duration).Format(format)
		value, hash := derive(seed, period, length, mode)
		slugs[i] = Slug{Value: value, Period: period, Hash: hash}
	}
	return slugs, nil
}

func derive(seed, period string, length int, mode string) (string, string) {
	entropy := hmacSHA256(seed, seed+":"+period)

	if strings.EqualFold(mode, "obfuscated") {
		slug := buildObfuscatedSlug(entropy)
		hash := hmacSHA256(seed, seed+":skid:"+period)
		hashLen := min((length+1)/2, 16)
		return slug, hex.EncodeToString(hash[:hashLen])
	}

	// BIP39 mode: concatenate mnemonic words
	words := entropyToBIP39Words(entropy)
	wordCount := min(length, 24)
	hashLen := (wordCount*11 + 7) / 8
	return strings.Join(words[:wordCount], ""), hex.EncodeToString(entropy[:hashLen])
}

func hmacSHA256(key, message string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return h.Sum(nil)
}

func entropyToBIP39Words(entropy []byte) []string {
	// Standard BIP39: 256 bits entropy + 8 bits checksum = 264 bits = 24 words
	checksum := sha256.Sum256(entropy)
	allBits := slices.Concat(bytesToBits(entropy), bytesToBits(checksum[:1]))

	words := make([]string, 24)
	for i := range words {
		index := bitsToInt(allBits[i*11 : i*11+11])
		words[i] = bip39Words[index]
	}
	return words
}

func bytesToBits(data []byte) []byte {
	bits := make([]byte, len(data)*8)
	for i, b := range data {
		for j := range 8 {
			bits[i*8+j] = (b >> (7 - j)) & 1
		}
	}
	return bits
}

func bitsToInt(bits []byte) int {
	val := 0
	for _, b := range bits {
		val = val<<1 | int(b)
	}
	return val
}

// Word lists for obfuscated mode
var (
	consonants = []string{"b", "c", "d", "f", "g", "k", "l", "m", "n", "p", "r", "s", "t", "v", "z"}
	vowels     = []string{"a", "a", "e", "e", "i", "i", "o", "o", "u"}
	codas      = []string{"", "", "", "", "", "", "n", "m", "r", "x"}
	prefixes   = []string{"get", "try", "go", "my", "pro", "on", "up", "hi"}
	suffixes   = []string{"ly", "fy", "io", "co", "go", "up", "hq", "ai"}
	numbers    = []string{"1", "2", "3", "4", "5", "7", "8", "9", "11", "22", "24", "42", "99", "101", "123", "247", "360", "365"}
	techWords  = []string{
		"cloud", "data", "tech", "sync", "fast", "smart", "link", "soft", "core", "base",
		"meta", "flux", "grid", "node", "edge", "wave", "pixel", "cyber", "logic", "delta",
		"sigma", "alpha", "beta", "gamma", "nova", "nexus", "pulse", "spark", "beam", "volt",
		"zero", "next", "snap", "dash", "rush", "bolt", "jump", "flip", "spin", "zoom",
		"push", "pull", "grab", "drop", "lift", "kick", "click", "swipe", "pure", "bold",
		"keen", "swift", "prime", "peak", "true", "safe", "bright", "clear", "clean", "fresh",
		"sharp", "super", "ultra", "mega", "rock", "star", "moon", "sand", "leaf", "pine",
		"oak", "wolf", "lake", "river", "wind", "fire", "ice", "snow", "rain", "sun",
		"fox", "bear", "hawk", "crow", "elk", "owl", "lion", "tiger", "blue", "red",
		"gray", "gold", "jade", "mint", "rust", "onyx", "amber", "coral", "ivory", "slate",
		"steel", "silver", "copper", "box", "hub", "lab", "bit", "dot", "max", "zen",
		"arc", "top", "pop", "cup", "cap", "pin", "pen", "pad", "pod",
	}
	blockedWords = []string{
		"shit", "fuck", "damn", "hell", "crap", "piss", "cock", "dick", "cunt", "ass",
		"fag", "nig", "sex", "xxx", "porn", "anal", "rape", "kill", "nazi", "hate",
		"dead", "die", "hack", "crack",
	}
)

// pick selects from choices using entropy byte at offset, returns selection and next offset
func pick(entropy []byte, offset int, choices []string) (string, int) {
	index := int(entropy[offset%32]) % len(choices)
	return choices[index], offset + 1
}

// makeSyllable creates a consonant-vowel-coda pattern like "ba", "kem", "tor"
func makeSyllable(entropy []byte, offset int) (string, int) {
	c, offset := pick(entropy, offset, consonants)
	v, offset := pick(entropy, offset, vowels)
	d, offset := pick(entropy, offset, codas)
	return c + v + d, offset
}

// shortenWord creates startup-style shortened names: tiger→tigr, delta→delt
func shortenWord(word string) string {
	n := len(word)
	if n < 4 {
		return word
	}
	// tiger → tigr (vowel before final r)
	if strings.ContainsAny(string(word[n-2]), "aeo") && word[n-1] == 'r' {
		return word[:n-2] + "r"
	}
	// purple → purpl (remove final e from -le)
	if strings.HasSuffix(word, "le") && n > 3 {
		return word[:n-1]
	}
	// delta → delt (remove trailing vowel)
	if strings.ContainsAny(string(word[n-1]), "aeiou") && n > 4 {
		return word[:n-1]
	}
	return word
}

func containsBlockedWord(s string) bool {
	lower := strings.ToLower(s)
	for _, word := range blockedWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

// buildObfuscatedSlug creates a startup-style name like "trybeambold8"
// Structure: [prefix] + word1 + [mid] + word2 + ending
func buildObfuscatedSlug(entropy []byte) string {
	const minLen, maxLen = 10, 18
	offset := 0
	var slug strings.Builder

	// 25% chance of prefix
	if entropy[offset]%4 == 0 {
		prefix, next := pick(entropy, offset+1, prefixes)
		slug.WriteString(prefix)
		offset = next
	} else {
		offset++
	}

	// First word (20% chance to shorten)
	word1, offset := pick(entropy, offset, techWords)
	if entropy[offset]%5 == 0 {
		word1 = shortenWord(word1)
	}
	offset++
	slug.WriteString(word1)

	// 15% chance of mid element (syllable or dash)
	if entropy[offset]%7 < 2 {
		if entropy[offset]%2 == 0 {
			mid, next := makeSyllable(entropy, offset+1)
			slug.WriteString(mid)
			offset = next
		} else {
			slug.WriteString("-")
			offset++
		}
	} else {
		offset++
	}

	// Second word (must differ from first, 20% chance to shorten)
	var word2 string
	for range 5 {
		word2, offset = pick(entropy, offset, techWords)
		if word2 != word1 {
			break
		}
	}
	if entropy[offset]%5 == 0 {
		word2 = shortenWord(word2)
	}
	offset++
	slug.WriteString(word2)

	// Ending: syllable (37.5%), number (25%), double-syllable (25%), suffix (12.5%)
	switch entropy[offset] % 8 {
	case 0, 1, 2:
		syl, _ := makeSyllable(entropy, offset+1)
		slug.WriteString(syl)
	case 3, 4:
		num, _ := pick(entropy, offset+1, numbers)
		slug.WriteString(num)
	case 5, 6:
		syl1, next := makeSyllable(entropy, offset+1)
		syl2, _ := makeSyllable(entropy, next)
		slug.WriteString(syl1 + syl2)
	case 7:
		suf, _ := pick(entropy, offset+1, suffixes)
		slug.WriteString(suf)
	}

	result := slug.String()
	result = padToMinLength(result, minLen, entropy)
	result = removeBlockedWords(result, entropy)
	result = truncateToMaxLength(result, minLen, maxLen)
	result = removeTripleLetters(result)
	return result
}

func padToMinLength(s string, minLen int, entropy []byte) string {
	offset := 20
	for len(s) < minLen {
		syl, next := makeSyllable(entropy, offset)
		s += syl
		offset = next
	}
	return s
}

func removeBlockedWords(s string, entropy []byte) string {
	for i := 0; containsBlockedWord(s) && i < 10; i++ {
		lower := strings.ToLower(s)
		for _, blocked := range blockedWords {
			if idx := strings.Index(lower, blocked); idx >= 0 {
				insertAt := idx + len(blocked)/2
				syl, _ := makeSyllable(entropy, 25+i)
				s = s[:insertAt] + syl + s[insertAt:]
				break
			}
		}
	}
	return s
}

func truncateToMaxLength(s string, minLen, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Prefer cutting at vowel boundary
	for i := maxLen; i >= minLen; i-- {
		if strings.ContainsAny(string(s[i-1]), "aeiou") {
			return s[:i]
		}
	}
	return s[:maxLen]
}

func removeTripleLetters(s string) string {
	var result strings.Builder
	for i, c := range s {
		if i < 2 || !(byte(c) == s[i-1] && byte(c) == s[i-2]) {
			result.WriteByte(byte(c))
		}
	}
	return result.String()
}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02T15",
	"2006-01-02",
}

func parseTime(s string) (time.Time, error) {
	for _, format := range timeFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", s)
}

func parseInterval(s string) (time.Duration, string, error) {
	switch strings.ToLower(s) {
	case "s", "second", "seconds":
		return time.Second, "2006-01-02T15:04:05", nil
	case "m", "minute", "minutes":
		return time.Minute, "2006-01-02T15:04", nil
	case "h", "hour", "hours":
		return time.Hour, "2006-01-02T15", nil
	case "", "d", "day", "days":
		return 24 * time.Hour, "2006-01-02", nil
	case "w", "week", "weeks":
		return 7 * 24 * time.Hour, "2006-W02", nil
	}
	return 0, "", fmt.Errorf("invalid interval: %s", s)
}
