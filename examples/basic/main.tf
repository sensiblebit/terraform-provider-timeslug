terraform {
  required_providers {
    timeslug = {
      source  = "sensiblebit/timeslug"
      version = "~> 0.0"
    }
  }
}

variable "seed" {
  description = "Secret seed for slug generation"
  type        = string
  sensitive   = true
}

provider "timeslug" {
  seed = var.seed
}

# BIP39 mode - concatenated mnemonic words
data "timeslug_slugs" "bip39_daily" {
  anchor   = "2026-02-03"
  length   = 3
  window   = 7
  interval = "day"
  mode     = "bip39"
}

# Obfuscated mode - startup-style slugs
data "timeslug_slugs" "obfuscated_daily" {
  anchor   = "2026-02-03"
  length   = 16
  window   = 7
  interval = "day"
  mode     = "obfuscated"
}

output "bip39_slugs" {
  description = "BIP39 word-based slugs"
  value       = data.timeslug_slugs.bip39_daily.slugs
}

output "obfuscated_slugs" {
  description = "Startup-style slugs"
  value       = data.timeslug_slugs.obfuscated_daily.slugs
}

output "current_bip39" {
  description = "Current BIP39 slug (center of window)"
  value       = data.timeslug_slugs.bip39_daily.slugs[3].slug
}

output "current_obfuscated" {
  description = "Current obfuscated slug (center of window)"
  value       = data.timeslug_slugs.obfuscated_daily.slugs[3].slug
}
