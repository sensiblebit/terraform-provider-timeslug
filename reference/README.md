# TimeSlug Reference Implementations

Reference implementations of the TimeSlug algorithm in multiple languages.
All implementations produce identical output for the same inputs.

## Modes

- **bip39**: Generates concatenated BIP39 words (e.g., "exoticangryanswer")
- **obfuscated**: Generates startup-style slugs (e.g., "trybeambold8")

## Usage

### Python

```bash
python3 timeslug.py <seed> <period> <mode> <length>
python3 timeslug.py seedphrase 2026-02-03 obfuscated 16
python3 timeslug.py seedphrase 2026-02-03 bip39 3
```

### Java

```bash
javac TimeSlug.java
java TimeSlug <seed> <period> <mode> <length>
java TimeSlug seedphrase 2026-02-03 obfuscated 16
```

### C++

```bash
# macOS with Homebrew OpenSSL
g++ -std=c++17 -O2 -o timeslug timeslug.cpp \
    -I$(brew --prefix openssl)/include \
    -L$(brew --prefix openssl)/lib -lcrypto

# Linux
g++ -std=c++17 -O2 -o timeslug timeslug.cpp -lcrypto

./timeslug <seed> <period> <mode> <length>
./timeslug seedphrase 2026-02-03 obfuscated 16
```

## Test Vectors

All implementations must produce these exact outputs:

| Seed | Period | Mode | Length | Slug | Hash |
|------|--------|------|--------|------|------|
| seedphrase | 2026-02-03 | obfuscated | 16 | trybeambold8 | 5d3bf0d55db67ea2 |
| seedphrase | 2026-02-04 | obfuscated | 16 | brightbeamvivar | f9fb66a05050f52f |
| seedphrase | 2026-02-05 | obfuscated | 16 | trycorefastfum | 8bb68bd056e4a6ff |
| seedphrase | 2026-02-03 | bip39 | 3 | exoticangryanswer | 50011c26d0 |
| seedphrase | 2026-02-03 | bip39 | 5 | exoticangryanswerpatternmain | 50011c26d0a864 |

## Algorithm Overview

### Obfuscated Mode (Layered Construction)

Structure: `[prefix] + word1 + [mid] + word2 + ending`

1. **Layer 1**: Optional prefix (25% chance)
2. **Layer 2**: First word (20% chance to shorten)
3. **Layer 3**: Optional mid element (15% chance: syllable or dash)
4. **Layer 4**: Second word, distinct from first (20% chance to shorten)
5. **Layer 5**: Ending (syllable 37.5%, number 25%, double-syllable 25%, suffix 12.5%)

### BIP39 Mode

1. HMAC-SHA256(seed, seed + ":" + period) → 32 bytes entropy
2. SHA256 checksum of entropy
3. Concatenate 256 entropy bits + 8 checksum bits
4. Convert to 24 x 11-bit indices
5. Map indices to BIP39 wordlist
