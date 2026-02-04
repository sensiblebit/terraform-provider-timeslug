---
page_title: "timeslug_slugs Data Source - terraform-provider-timeslug"
subcategory: ""
description: |-
  Generates deterministic slugs for a rolling time window.
---

# timeslug_slugs (Data Source)

Generates deterministic slugs for a rolling time window centered on an anchor time. Each slug includes the value, time period, and a verification hash.

## Example Usage

### BIP39 Mode (Default)

```terraform
data "timeslug_slugs" "daily" {
  anchor   = "2026-02-03"
  length   = 3      # 3 BIP39 words
  window   = 7      # 7 days
  interval = "day"
  mode     = "bip39"
}

# Output: exoticangryanswer
output "today" {
  value = data.timeslug_slugs.daily.slugs[3].slug
}
```

### Obfuscated Mode

```terraform
data "timeslug_slugs" "hourly" {
  anchor   = "2026-02-03T12:00:00"
  length   = 16     # Target character length
  window   = 5      # 5 hours
  interval = "hour"
  mode     = "obfuscated"
}

# Output: trybeambold8
output "current_hour" {
  value = data.timeslug_slugs.hourly.slugs[2].slug
}
```

### All Intervals

```terraform
# Second-level rotation
data "timeslug_slugs" "seconds" {
  anchor   = "2026-02-03T12:30:45"
  interval = "second"
  window   = 10
}

# Minute-level rotation
data "timeslug_slugs" "minutes" {
  anchor   = "2026-02-03T12:30"
  interval = "minute"
  window   = 10
}

# Weekly rotation
data "timeslug_slugs" "weeks" {
  anchor   = "2026-02-03"
  interval = "week"
  window   = 4
}
```

## Schema

### Required

- `anchor` (String) Center point for the time window. Supported formats:
  - `2006-01-02` (date only)
  - `2006-01-02T15` (with hour)
  - `2006-01-02T15:04` (with minute)
  - `2006-01-02T15:04:05` (with second)
  - RFC3339 format

### Optional

- `length` (Number) Slug length. For `bip39` mode: number of words (1-24). For `obfuscated` mode: target character length. Default: `3`
- `window` (Number) Number of periods in the window. Default: `7`
- `interval` (String) Rotation interval. One of: `second`, `minute`, `hour`, `day`, `week`. Default: `day`
- `mode` (String) Output mode. One of: `bip39`, `obfuscated`. Default: `bip39`

### Read-Only

- `id` (String) Unique identifier for this data source configuration.
- `slugs` (List of Object) Generated slugs for the time window. Each object contains:
  - `slug` (String) The generated slug value.
  - `period` (String) The time period this slug is valid for.
  - `hash` (String) Verification hash for this slug.

## Modes

### BIP39 Mode

Generates slugs by concatenating BIP39 mnemonic words. The slug is derived using:

1. HMAC-SHA256 of `seed:period` to generate 32 bytes of entropy
2. SHA256 checksum appended to create 264 bits
3. Split into 24 x 11-bit indices
4. Map to BIP39 wordlist (2048 words)

Example outputs:
- 3 words: `exoticangryanswer`
- 5 words: `exoticangryanswerpatternmain`

### Obfuscated Mode

Generates startup-style slugs using layered construction:

- Optional prefix: `get`, `try`, `go`, `my`, `pro`, `on`, `up`, `hi`
- Two tech/nature words: `cloud`, `beam`, `fast`, `bold`, etc.
- Ending: syllable, number, or suffix

Example outputs: `trybeambold8`, `brightbeamvivar`, `trycorefastfum`
