import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.MessageDigest;
import java.util.*;

/**
 * TimeSlug Reference Implementation - Java
 * Generates deterministic slugs in two modes: BIP39 (words) and Obfuscated (synth)
 */
public class TimeSlug {
    // Synth constants
    private static final String[] CONSONANTS = {"b", "c", "d", "f", "g", "k", "l", "m", "n", "p", "r", "s", "t", "v", "z"};
    private static final String[] VOWELS = {"a", "a", "e", "e", "i", "i", "o", "o", "u"};
    private static final String[] CODAS = {"", "", "", "", "", "", "n", "m", "r", "x"};
    private static final String[] PREFIXES = {"get", "try", "go", "my", "pro", "on", "up", "hi"};
    private static final String[] SUFFIXES = {"ly", "fy", "io", "co", "go", "up", "hq", "ai"};
    private static final String[] NUMBERS = {"1", "2", "3", "4", "5", "7", "8", "9", "11", "22", "24", "42", "99", "101", "123", "247", "360", "365"};
    private static final String[] WORDS = {
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
    private static final String[] BLOCKED = {"shit", "fuck", "damn", "hell", "crap", "piss", "cock", "dick", "cunt", "ass",
        "fag", "nig", "sex", "xxx", "porn", "anal", "rape", "kill", "nazi", "hate", "dead", "die", "hack", "crack"};

    private static List<String> bip39Words;
    private static int offset;

    public static void main(String[] args) throws Exception {
        // Load BIP39 wordlist
        String bip39Path = "../internal/provider/bip39_english.txt";
        bip39Words = Files.readAllLines(Path.of(bip39Path));

        String seed = args.length > 0 ? args[0] : "seedphrase";
        String period = args.length > 1 ? args[1] : "2026-02-03";
        String mode = args.length > 2 ? args[2] : "obfuscated";
        int length = args.length > 3 ? Integer.parseInt(args[3]) : 16;

        String[] result = derive(seed, period, length, mode);
        System.out.println("Mode:   " + mode);
        System.out.println("Period: " + period);
        System.out.println("Slug:   " + result[0]);
        System.out.println("Hash:   " + result[1]);
    }

    static byte[] hmacHash(String key, String message) throws Exception {
        Mac mac = Mac.getInstance("HmacSHA256");
        mac.init(new SecretKeySpec(key.getBytes(), "HmacSHA256"));
        return mac.doFinal(message.getBytes());
    }

    static String pick(byte[] entropy, String[] choices) {
        int idx = (entropy[offset % 32] & 0xFF) % choices.length;
        offset++;
        return choices[idx];
    }

    static String syllable(byte[] entropy) {
        return pick(entropy, CONSONANTS) + pick(entropy, VOWELS) + pick(entropy, CODAS);
    }

    static String shorten(String word) {
        int n = word.length();
        if (n < 4) return word;
        if (n >= 3 && "aeo".indexOf(word.charAt(n - 2)) >= 0 && word.charAt(n - 1) == 'r')
            return word.substring(0, n - 2) + "r";
        if (word.endsWith("le") && n > 3)
            return word.substring(0, n - 1);
        if ("aeiou".indexOf(word.charAt(n - 1)) >= 0 && n > 4)
            return word.substring(0, n - 1);
        return word;
    }

    static boolean hasBlocked(String s) {
        String lower = s.toLowerCase();
        for (String b : BLOCKED) if (lower.contains(b)) return true;
        return false;
    }

    static String buildSynth(byte[] entropy) {
        final int MIN_LEN = 10, MAX_LEN = 18;
        offset = 0;
        StringBuilder result = new StringBuilder();

        // Layer 1: Optional prefix (25%)
        if ((entropy[offset] & 0xFF) % 4 == 0) {
            offset++;
            result.append(pick(entropy, PREFIXES));
        } else {
            offset++;
        }

        // Layer 2: First word (20% shorten)
        String w1 = pick(entropy, WORDS);
        if ((entropy[offset] & 0xFF) % 5 == 0) w1 = shorten(w1);
        offset++;
        result.append(w1);

        // Layer 3: Optional mid (15%)
        if ((entropy[offset] & 0xFF) % 7 < 2) {
            if ((entropy[offset] & 0xFF) % 2 == 0) {
                offset++;
                result.append(syllable(entropy));
            } else {
                result.append("-");
                offset++;
            }
        } else {
            offset++;
        }

        // Layer 4: Second word (20% shorten)
        String w2 = w1;
        for (int i = 0; i < 5 && w2.equals(w1); i++) {
            w2 = pick(entropy, WORDS);
        }
        if ((entropy[offset] & 0xFF) % 5 == 0) w2 = shorten(w2);
        offset++;
        result.append(w2);

        // Layer 5: Ending
        int et = (entropy[offset] & 0xFF) % 8;
        offset++;
        if (et <= 2) {
            result.append(syllable(entropy));
        } else if (et <= 4) {
            result.append(pick(entropy, NUMBERS));
        } else if (et <= 6) {
            result.append(syllable(entropy)).append(syllable(entropy));
        } else {
            result.append(pick(entropy, SUFFIXES));
        }

        String slug = result.toString();

        // Pad
        int po = 20;
        while (slug.length() < MIN_LEN) {
            offset = po;
            slug += syllable(entropy);
            po = offset;
        }

        // Fix blocked
        for (int i = 0; hasBlocked(slug) && i < 10; i++) {
            for (String b : BLOCKED) {
                int idx = slug.toLowerCase().indexOf(b);
                if (idx >= 0) {
                    int insertAt = idx + b.length() / 2;
                    offset = 25 + i;
                    slug = slug.substring(0, insertAt) + syllable(entropy) + slug.substring(insertAt);
                    break;
                }
            }
        }

        // Truncate
        if (slug.length() > MAX_LEN) {
            for (int i = MAX_LEN; i >= MIN_LEN; i--) {
                if ("aeiou".indexOf(slug.charAt(i - 1)) >= 0) {
                    slug = slug.substring(0, i);
                    break;
                }
            }
            if (slug.length() > MAX_LEN) slug = slug.substring(0, MAX_LEN);
        }

        // Remove triple letters
        StringBuilder cleaned = new StringBuilder();
        for (int i = 0; i < slug.length(); i++) {
            char c = slug.charAt(i);
            if (i < 2 || !(c == slug.charAt(i - 1) && c == slug.charAt(i - 2))) {
                cleaned.append(c);
            }
        }

        return cleaned.toString();
    }

    static String[] entropyToWords(byte[] entropy) throws Exception {
        byte[] checksum = MessageDigest.getInstance("SHA-256").digest(entropy);

        // Build 264 bits
        int[] bits = new int[264];
        for (int i = 0; i < 32; i++) {
            for (int j = 0; j < 8; j++) {
                bits[i * 8 + j] = (entropy[i] >> (7 - j)) & 1;
            }
        }
        for (int j = 0; j < 8; j++) {
            bits[256 + j] = (checksum[0] >> (7 - j)) & 1;
        }

        // Convert to words
        String[] words = new String[24];
        for (int i = 0; i < 24; i++) {
            int idx = 0;
            for (int j = 0; j < 11; j++) {
                idx = (idx << 1) | bits[i * 11 + j];
            }
            words[i] = bip39Words.get(idx);
        }
        return words;
    }

    static String[] derive(String seed, String period, int length, String mode) throws Exception {
        byte[] entropy = hmacHash(seed, seed + ":" + period);

        if (mode.equalsIgnoreCase("obfuscated")) {
            String value = buildSynth(entropy);
            byte[] altHash = hmacHash(seed, seed + ":skid:" + period);
            int hashLen = Math.min((length + 1) / 2, 16);
            return new String[]{value, bytesToHex(altHash, hashLen)};
        }

        // BIP39 mode
        String[] words = entropyToWords(entropy);
        int wordCount = Math.min(length, 24);
        int hashLen = (wordCount * 11 + 7) / 8;
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < wordCount; i++) sb.append(words[i]);
        return new String[]{sb.toString(), bytesToHex(entropy, hashLen)};
    }

    static String bytesToHex(byte[] bytes, int len) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < len; i++) sb.append(String.format("%02x", bytes[i]));
        return sb.toString();
    }
}
