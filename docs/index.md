---
page_title: "Provider: TimeSlug"
description: |-
  The TimeSlug provider generates deterministic time-rotating slugs for URLs and identifiers.
---

# TimeSlug Provider

The TimeSlug provider generates deterministic, time-rotating slugs from a secret seed. Slugs rotate on configurable intervals (seconds to weeks) and can be generated in two modes:

- **BIP39**: Concatenated BIP39 mnemonic words (e.g., `exoticangryanswer`)
- **Obfuscated**: Startup-style alphanumeric slugs (e.g., `trybeambold8`)

## Use Cases

- **Time-based tokens** that are predictable for authorized parties
- **Deterministic identifiers** that change on schedule
- **Rotating secrets** for URLs or API endpoints

## Example Usage

```terraform
terraform {
  required_providers {
    timeslug = {
      source  = "sensiblebit/timeslug"
      version = "~> 1.0"
    }
  }
}

provider "timeslug" {
  seed = var.secret_seed  # Keep this secret!
}

data "timeslug_slugs" "rotating" {
  anchor   = "2026-02-03"
  length   = 3
  window   = 7
  interval = "day"
  mode     = "bip39"
}

output "current_slug" {
  value = data.timeslug_slugs.rotating.slugs[3].slug  # Center of window
}

output "all_slugs" {
  value = data.timeslug_slugs.rotating.slugs
}
```

## Schema

### Required

- `seed` (String, Sensitive) Secret seed for slug generation. All slugs are deterministically derived from this value.

## Security Notes

- The `seed` is marked as sensitive and will not appear in logs or state output
- Anyone with the seed can predict all past and future slugs
- Use a strong, random seed value
