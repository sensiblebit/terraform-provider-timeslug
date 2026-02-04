# Terraform Provider: TimeSlug

Generate deterministic, time-rotating slugs for URLs and identifiers.

## Features

- **Deterministic**: Same seed + time period = same slug, every time
- **Time-rotating**: Slugs change on configurable intervals (seconds to weeks)
- **Two modes**:
  - `bip39`: Concatenated BIP39 mnemonic words (`exoticangryanswer`)
  - `obfuscated`: Startup-style alphanumeric slugs (`trybeambold8`)
- **Cross-platform**: Reference implementations in Go, Python, Java, and C++

## Installation

```terraform
terraform {
  required_providers {
    timeslug = {
      source  = "sensiblebit/timeslug"
      version = "~> 0.0"
    }
  }
}

provider "timeslug" {
  seed = var.secret_seed
}
```

## Usage

```terraform
# Generate daily rotating slugs
data "timeslug_slugs" "daily" {
  anchor   = "2026-02-03"
  length   = 3
  window   = 7
  interval = "day"
  mode     = "bip39"
}

# Output the current slug
output "current_slug" {
  value = data.timeslug_slugs.daily.slugs[3].slug
}
```

## Data Sources

### timeslug_slugs

Generates slugs for a rolling time window.

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `anchor` | string | yes | - | Center time for the window |
| `length` | number | no | 3 | Words (bip39) or chars (obfuscated) |
| `window` | number | no | 7 | Number of periods |
| `interval` | string | no | day | second, minute, hour, day, week |
| `mode` | string | no | bip39 | bip39 or obfuscated |

#### Output

```hcl
slugs = [
  { slug = "...", period = "2026-02-01", hash = "..." },
  { slug = "...", period = "2026-02-02", hash = "..." },
  ...
]
```

## Test Vectors

All implementations produce identical output:

| Seed | Period | Mode | Length | Slug | Hash |
|------|--------|------|--------|------|------|
| seedphrase | 2026-02-03 | obfuscated | 16 | trybeambold8 | 5d3bf0d55db67ea2 |
| seedphrase | 2026-02-04 | obfuscated | 16 | brightbeamvivar | f9fb66a05050f52f |
| seedphrase | 2026-02-05 | obfuscated | 16 | trycorefastfum | 8bb68bd056e4a6ff |
| seedphrase | 2026-02-03 | bip39 | 3 | exoticangryanswer | 50011c26d0 |
| seedphrase | 2026-02-03 | bip39 | 5 | exoticangryanswerpatternmain | 50011c26d0a864 |

## Reference Implementations

See the `reference/` directory for implementations in:

- **Python**: `python3 timeslug.py <seed> <period> <mode> <length>`
- **Java**: `java TimeSlug <seed> <period> <mode> <length>`
- **C++**: `./timeslug <seed> <period> <mode> <length>`

## Building

```bash
go build -o terraform-provider-timeslug
```

## Testing

```bash
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE)
