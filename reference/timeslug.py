#!/usr/bin/env python3
"""
TimeSlug Reference Implementation - Python
Generates deterministic slugs in two modes: BIP39 (words) and Obfuscated (synth)
"""

import hashlib
import hmac
import sys

# BIP39 wordlist (first 20 shown, full list in production)
BIP39_WORDS = open('bip39_english.txt').read().strip().split('\n') if __name__ != '__main__' else None

# Synth constants
CONSONANTS = ['b', 'c', 'd', 'f', 'g', 'k', 'l', 'm', 'n', 'p', 'r', 's', 't', 'v', 'z']
VOWELS = ['a', 'a', 'e', 'e', 'i', 'i', 'o', 'o', 'u']
CODAS = ['', '', '', '', '', '', 'n', 'm', 'r', 'x']
PREFIXES = ['get', 'try', 'go', 'my', 'pro', 'on', 'up', 'hi']
SUFFIXES = ['ly', 'fy', 'io', 'co', 'go', 'up', 'hq', 'ai']
NUMBERS = ['1', '2', '3', '4', '5', '7', '8', '9', '11', '22', '24', '42', '99', '101', '123', '247', '360', '365']
WORDS = [
    'cloud', 'data', 'tech', 'sync', 'fast', 'smart', 'link', 'soft', 'core', 'base',
    'meta', 'flux', 'grid', 'node', 'edge', 'wave', 'pixel', 'cyber', 'logic', 'delta',
    'sigma', 'alpha', 'beta', 'gamma', 'nova', 'nexus', 'pulse', 'spark', 'beam', 'volt',
    'zero', 'next', 'snap', 'dash', 'rush', 'bolt', 'jump', 'flip', 'spin', 'zoom',
    'push', 'pull', 'grab', 'drop', 'lift', 'kick', 'click', 'swipe', 'pure', 'bold',
    'keen', 'swift', 'prime', 'peak', 'true', 'safe', 'bright', 'clear', 'clean', 'fresh',
    'sharp', 'super', 'ultra', 'mega', 'rock', 'star', 'moon', 'sand', 'leaf', 'pine',
    'oak', 'wolf', 'lake', 'river', 'wind', 'fire', 'ice', 'snow', 'rain', 'sun',
    'fox', 'bear', 'hawk', 'crow', 'elk', 'owl', 'lion', 'tiger', 'blue', 'red',
    'gray', 'gold', 'jade', 'mint', 'rust', 'onyx', 'amber', 'coral', 'ivory', 'slate',
    'steel', 'silver', 'copper', 'box', 'hub', 'lab', 'bit', 'dot', 'max', 'zen',
    'arc', 'top', 'pop', 'cup', 'cap', 'pin', 'pen', 'pad', 'pod',
]
BLOCKED = ['shit', 'fuck', 'damn', 'hell', 'crap', 'piss', 'cock', 'dick', 'cunt', 'ass',
           'fag', 'nig', 'sex', 'xxx', 'porn', 'anal', 'rape', 'kill', 'nazi', 'hate',
           'dead', 'die', 'hack', 'crack']


def hmac_hash(key: str, message: str) -> bytes:
    """Generate HMAC-SHA256."""
    return hmac.new(key.encode(), message.encode(), hashlib.sha256).digest()


def pick(entropy: bytes, offset: int, choices: list):
    """Pick item from list using entropy byte."""
    return choices[entropy[offset % 32] % len(choices)], offset + 1


def syllable(entropy: bytes, offset: int):
    """Generate consonant-vowel-coda syllable."""
    c, offset = pick(entropy, offset, CONSONANTS)
    v, offset = pick(entropy, offset, VOWELS)
    d, offset = pick(entropy, offset, CODAS)
    return c + v + d, offset


def shorten(word: str) -> str:
    """Startup-style shortening: tiger->tigr, delta->delt."""
    n = len(word)
    if n < 4:
        return word
    if n >= 3 and word[-2] in 'aeo' and word[-1] == 'r':
        return word[:-2] + 'r'
    if word.endswith('le') and n > 3:
        return word[:-1]
    if word[-1] in 'aeiou' and n > 4:
        return word[:-1]
    return word


def has_blocked(s: str) -> bool:
    """Check for blocked words."""
    lower = s.lower()
    return any(b in lower for b in BLOCKED)


def build_synth(entropy: bytes) -> str:
    """Build obfuscated slug using layered construction."""
    MIN_LEN, MAX_LEN = 10, 18
    o = 0
    result = []

    # Layer 1: Optional prefix (25%)
    if entropy[o] % 4 == 0:
        p, o = pick(entropy, o + 1, PREFIXES)
        result.append(p)
    else:
        o += 1

    # Layer 2: First word (20% shorten)
    w1, o = pick(entropy, o, WORDS)
    if entropy[o] % 5 == 0:
        w1 = shorten(w1)
    o += 1
    result.append(w1)

    # Layer 3: Optional mid (15%)
    if entropy[o] % 7 < 2:
        if entropy[o] % 2 == 0:
            mid, o = syllable(entropy, o + 1)
            result.append(mid)
        else:
            result.append('-')
            o += 1
    else:
        o += 1

    # Layer 4: Second word (20% shorten)
    for _ in range(5):
        w2, o = pick(entropy, o, WORDS)
        if w2 != w1:
            break
    if entropy[o] % 5 == 0:
        w2 = shorten(w2)
    o += 1
    result.append(w2)

    # Layer 5: Ending
    et = entropy[o] % 8
    if et <= 2:
        syl, _ = syllable(entropy, o + 1)
        result.append(syl)
    elif et <= 4:
        num, _ = pick(entropy, o + 1, NUMBERS)
        result.append(num)
    elif et <= 6:
        s1, no = syllable(entropy, o + 1)
        s2, _ = syllable(entropy, no)
        result.append(s1 + s2)
    else:
        suf, _ = pick(entropy, o + 1, SUFFIXES)
        result.append(suf)

    slug = ''.join(result)

    # Pad
    po = 20
    while len(slug) < MIN_LEN:
        syl, po = syllable(entropy, po)
        slug += syl

    # Fix blocked
    for i in range(10):
        if not has_blocked(slug):
            break
        for b in BLOCKED:
            if b in slug.lower():
                idx = slug.lower().index(b) + len(b) // 2
                syl, _ = syllable(entropy, 25 + i)
                slug = slug[:idx] + syl + slug[idx:]
                break

    # Truncate
    if len(slug) > MAX_LEN:
        for i in range(MAX_LEN, MIN_LEN - 1, -1):
            if slug[i-1] in 'aeiou':
                slug = slug[:i]
                break
        else:
            slug = slug[:MAX_LEN]

    # Remove triple letters
    cleaned = []
    for i, c in enumerate(slug):
        if i < 2 or not (c == slug[i-1] == slug[i-2]):
            cleaned.append(c)

    return ''.join(cleaned)


def entropy_to_words(entropy: bytes, bip39_words: list) -> list:
    """Convert 32 bytes entropy to 24 BIP39 words."""
    checksum = hashlib.sha256(entropy).digest()
    
    # Build 264 bits: 256 entropy + 8 checksum
    bits = []
    for b in entropy:
        for j in range(8):
            bits.append((b >> (7 - j)) & 1)
    for j in range(8):
        bits.append((checksum[0] >> (7 - j)) & 1)

    # Convert 11-bit chunks to word indices
    words = []
    for i in range(24):
        idx = 0
        for j in range(11):
            idx = (idx << 1) | bits[i * 11 + j]
        words.append(bip39_words[idx])
    
    return words


def derive(seed: str, period: str, length: int, mode: str, bip39_words: list = None):
    """Generate slug and hash for given seed/period."""
    entropy = hmac_hash(seed, f"{seed}:{period}")

    if mode.lower() == 'obfuscated':
        value = build_synth(entropy)
        alt_hash = hmac_hash(seed, f"{seed}:skid:{period}")
        hash_len = min((length + 1) // 2, 16)
        return value, alt_hash[:hash_len].hex()

    # BIP39 mode
    words = entropy_to_words(entropy, bip39_words)
    word_count = min(length, 24)
    hash_len = (word_count * 11 + 7) // 8
    return ''.join(words[:word_count]), entropy[:hash_len].hex()


if __name__ == '__main__':
    # Load BIP39 wordlist
    import os
    script_dir = os.path.dirname(os.path.abspath(__file__))
    bip39_path = os.path.join(script_dir, '..', 'internal', 'provider', 'bip39_english.txt')
    with open(bip39_path) as f:
        bip39_words = f.read().strip().split('\n')

    seed = sys.argv[1] if len(sys.argv) > 1 else 'seedphrase'
    period = sys.argv[2] if len(sys.argv) > 2 else '2026-02-03'
    mode = sys.argv[3] if len(sys.argv) > 3 else 'obfuscated'
    length = int(sys.argv[4]) if len(sys.argv) > 4 else 16

    slug, hash_val = derive(seed, period, length, mode, bip39_words)
    print(f"Mode:   {mode}")
    print(f"Period: {period}")
    print(f"Slug:   {slug}")
    print(f"Hash:   {hash_val}")
