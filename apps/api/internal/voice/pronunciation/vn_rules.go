package pronunciation

import (
	"strings"
	"unicode/utf8"
)

// CoachingTip maps one expected/heard IPA contrast to a short Vietnamese-
// learner coaching string. Static and cheap: no model, no per-request cost.
// Unknown contrasts return "" so the UI falls back to a generic prompt.
func CoachingTip(expected, heard string) string {
	e := strings.TrimSpace(expected)
	h := strings.TrimSpace(heard)
	if e == "" || h == "" || e == h {
		return ""
	}

	switch {
	case strings.Contains(e, "θ") && !strings.Contains(h, "θ") && strings.Contains(h, "t"):
		return "Put your tongue between your teeth for /θ/ (think), not /t/ (tink)."
	case strings.Contains(e, "ð") && !strings.Contains(h, "ð") && strings.Contains(h, "d"):
		return "Touch your tongue to your teeth and voice /ð/ (they), not /d/ (day)."
	case strings.Contains(e, "tʃ") && !strings.Contains(h, "tʃ") && strings.Contains(h, "tr"):
		return "Say /tʃ/ as one sound (chew), not /tr/ (true)."
	case strings.Contains(e, "ʃ") && !strings.Contains(h, "ʃ") && strings.Contains(h, "s"):
		return "Round your lips and pull the tongue back for /ʃ/ (she), not /s/ (see)."
	case strings.Contains(e, "l") && !strings.Contains(h, "l") && strings.Contains(h, "n"):
		return "Lift the tongue tip to the ridge for /l/ (light), not /n/ (night)."
	case strings.Contains(e, "ŋ") && !strings.Contains(h, "ŋ") && strings.Contains(h, "n"):
		return "End with /ŋ/ at the back of your mouth (sing), not /n/ (sin)."
	}

	if dropsFinalConsonant(e, h) {
		return "Say the final consonant — do not drop the end of the word."
	}
	return ""
}

// dropsFinalConsonant is the vowel-agnostic fallback: the reference ends in a
// consonant the learner's decode does not. It catches final -t/-k/-p and the
// -ed endings the minimal-pair list targets.
func dropsFinalConsonant(expected, heard string) bool {
	const consonants = "ptkbdgmnszlfrvθðʃʒŋ"
	last, _ := utf8.DecodeLastRuneInString(expected)
	if !strings.ContainsRune(consonants, last) {
		return false
	}
	return !strings.HasSuffix(heard, string(last))
}
