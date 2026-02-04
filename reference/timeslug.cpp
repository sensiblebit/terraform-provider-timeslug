/**
 * TimeSlug Reference Implementation - C++
 * Generates deterministic slugs in two modes: BIP39 (words) and Obfuscated (synth)
 * 
 * Compile: g++ -std=c++17 -O2 -o timeslug timeslug.cpp -lcrypto
 */

#include <openssl/hmac.h>
#include <openssl/sha.h>
#include <algorithm>
#include <cstring>
#include <fstream>
#include <iomanip>
#include <iostream>
#include <sstream>
#include <string>
#include <vector>

// Synth constants
const std::vector<std::string> CONSONANTS = {"b", "c", "d", "f", "g", "k", "l", "m", "n", "p", "r", "s", "t", "v", "z"};
const std::vector<std::string> VOWELS = {"a", "a", "e", "e", "i", "i", "o", "o", "u"};
const std::vector<std::string> CODAS = {"", "", "", "", "", "", "n", "m", "r", "x"};
const std::vector<std::string> PREFIXES = {"get", "try", "go", "my", "pro", "on", "up", "hi"};
const std::vector<std::string> SUFFIXES = {"ly", "fy", "io", "co", "go", "up", "hq", "ai"};
const std::vector<std::string> NUMBERS = {"1", "2", "3", "4", "5", "7", "8", "9", "11", "22", "24", "42", "99", "101", "123", "247", "360", "365"};
const std::vector<std::string> WORDS = {
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
    "arc", "top", "pop", "cup", "cap", "pin", "pen", "pad", "pod"
};
const std::vector<std::string> BLOCKED = {"shit", "fuck", "damn", "hell", "crap", "piss", "cock", "dick", "cunt", "ass",
    "fag", "nig", "sex", "xxx", "porn", "anal", "rape", "kill", "nazi", "hate", "dead", "die", "hack", "crack"};

std::vector<std::string> bip39Words;
int offset;

std::vector<unsigned char> hmacHash(const std::string& key, const std::string& message) {
    std::vector<unsigned char> result(32);
    unsigned int len = 32;
    HMAC(EVP_sha256(), key.c_str(), key.size(),
         reinterpret_cast<const unsigned char*>(message.c_str()), message.size(),
         result.data(), &len);
    return result;
}

std::string pick(const std::vector<unsigned char>& entropy, const std::vector<std::string>& choices) {
    int idx = entropy[offset % 32] % choices.size();
    offset++;
    return choices[idx];
}

std::string syllable(const std::vector<unsigned char>& entropy) {
    return pick(entropy, CONSONANTS) + pick(entropy, VOWELS) + pick(entropy, CODAS);
}

std::string shorten(const std::string& word) {
    size_t n = word.length();
    if (n < 4) return word;
    if (n >= 3 && std::string("aeo").find(word[n-2]) != std::string::npos && word[n-1] == 'r')
        return word.substr(0, n-2) + "r";
    if (word.length() > 3 && word.substr(n-2) == "le")
        return word.substr(0, n-1);
    if (std::string("aeiou").find(word[n-1]) != std::string::npos && n > 4)
        return word.substr(0, n-1);
    return word;
}

bool hasBlocked(const std::string& s) {
    std::string lower = s;
    std::transform(lower.begin(), lower.end(), lower.begin(), ::tolower);
    for (const auto& b : BLOCKED) {
        if (lower.find(b) != std::string::npos) return true;
    }
    return false;
}

std::string buildSynth(const std::vector<unsigned char>& entropy) {
    const int MIN_LEN = 10, MAX_LEN = 18;
    offset = 0;
    std::string result;

    // Layer 1: Optional prefix (25%)
    if (entropy[offset] % 4 == 0) {
        offset++;
        result += pick(entropy, PREFIXES);
    } else {
        offset++;
    }

    // Layer 2: First word (20% shorten)
    std::string w1 = pick(entropy, WORDS);
    if (entropy[offset] % 5 == 0) w1 = shorten(w1);
    offset++;
    result += w1;

    // Layer 3: Optional mid (15%)
    if (entropy[offset] % 7 < 2) {
        if (entropy[offset] % 2 == 0) {
            offset++;
            result += syllable(entropy);
        } else {
            result += "-";
            offset++;
        }
    } else {
        offset++;
    }

    // Layer 4: Second word (20% shorten)
    std::string w2 = w1;
    for (int i = 0; i < 5 && w2 == w1; i++) {
        w2 = pick(entropy, WORDS);
    }
    if (entropy[offset] % 5 == 0) w2 = shorten(w2);
    offset++;
    result += w2;

    // Layer 5: Ending
    int et = entropy[offset] % 8;
    offset++;
    if (et <= 2) {
        result += syllable(entropy);
    } else if (et <= 4) {
        result += pick(entropy, NUMBERS);
    } else if (et <= 6) {
        result += syllable(entropy) + syllable(entropy);
    } else {
        result += pick(entropy, SUFFIXES);
    }

    // Pad
    int po = 20;
    while ((int)result.length() < MIN_LEN) {
        offset = po;
        result += syllable(entropy);
        po = offset;
    }

    // Fix blocked
    for (int i = 0; hasBlocked(result) && i < 10; i++) {
        std::string lower = result;
        std::transform(lower.begin(), lower.end(), lower.begin(), ::tolower);
        for (const auto& b : BLOCKED) {
            size_t idx = lower.find(b);
            if (idx != std::string::npos) {
                int insertAt = idx + b.length() / 2;
                offset = 25 + i;
                result = result.substr(0, insertAt) + syllable(entropy) + result.substr(insertAt);
                break;
            }
        }
    }

    // Truncate
    if ((int)result.length() > MAX_LEN) {
        for (int i = MAX_LEN; i >= MIN_LEN; i--) {
            if (std::string("aeiou").find(result[i-1]) != std::string::npos) {
                result = result.substr(0, i);
                break;
            }
        }
        if ((int)result.length() > MAX_LEN) result = result.substr(0, MAX_LEN);
    }

    // Remove triple letters
    std::string cleaned;
    for (size_t i = 0; i < result.length(); i++) {
        char c = result[i];
        if (i < 2 || !(c == result[i-1] && c == result[i-2])) {
            cleaned += c;
        }
    }

    return cleaned;
}

std::vector<std::string> entropyToWords(const std::vector<unsigned char>& entropy) {
    unsigned char checksum[32];
    SHA256(entropy.data(), 32, checksum);

    // Build 264 bits
    std::vector<int> bits(264);
    for (int i = 0; i < 32; i++) {
        for (int j = 0; j < 8; j++) {
            bits[i * 8 + j] = (entropy[i] >> (7 - j)) & 1;
        }
    }
    for (int j = 0; j < 8; j++) {
        bits[256 + j] = (checksum[0] >> (7 - j)) & 1;
    }

    // Convert to words
    std::vector<std::string> words(24);
    for (int i = 0; i < 24; i++) {
        int idx = 0;
        for (int j = 0; j < 11; j++) {
            idx = (idx << 1) | bits[i * 11 + j];
        }
        words[i] = bip39Words[idx];
    }
    return words;
}

std::string bytesToHex(const std::vector<unsigned char>& bytes, int len) {
    std::ostringstream ss;
    for (int i = 0; i < len; i++) {
        ss << std::hex << std::setfill('0') << std::setw(2) << (int)bytes[i];
    }
    return ss.str();
}

std::pair<std::string, std::string> derive(const std::string& seed, const std::string& period, int length, const std::string& mode) {
    auto entropy = hmacHash(seed, seed + ":" + period);

    if (mode == "obfuscated") {
        std::string value = buildSynth(entropy);
        auto altHash = hmacHash(seed, seed + ":skid:" + period);
        int hashLen = std::min((length + 1) / 2, 16);
        return {value, bytesToHex(altHash, hashLen)};
    }

    // BIP39 mode
    auto words = entropyToWords(entropy);
    int wordCount = std::min(length, 24);
    int hashLen = (wordCount * 11 + 7) / 8;
    std::string slug;
    for (int i = 0; i < wordCount; i++) slug += words[i];
    return {slug, bytesToHex(entropy, hashLen)};
}

int main(int argc, char* argv[]) {
    // Load BIP39 wordlist
    std::ifstream file("../internal/provider/bip39_english.txt");
    std::string line;
    while (std::getline(file, line)) {
        bip39Words.push_back(line);
    }

    std::string seed = argc > 1 ? argv[1] : "seedphrase";
    std::string period = argc > 2 ? argv[2] : "2026-02-03";
    std::string mode = argc > 3 ? argv[3] : "obfuscated";
    int length = argc > 4 ? std::stoi(argv[4]) : 16;

    auto [slug, hash] = derive(seed, period, length, mode);
    std::cout << "Mode:   " << mode << std::endl;
    std::cout << "Period: " << period << std::endl;
    std::cout << "Slug:   " << slug << std::endl;
    std::cout << "Hash:   " << hash << std::endl;

    return 0;
}
