package pronunciation

import "sort"

// Confidence policy: the engine's per-word precision is low enough that
// showing every flag teaches the learner to distrust the panel. These are
// starting values, not validated policy. Calibration against the recorded
// fixtures (voice/tests/testdata, pinned by TestGatingRecordedFixtures): the
// fully-wrong attempt's word confidences are 0.749–1.0, so the word floor keeps
// all 7 and truncation shows 3; the phone floor drops 2 of 24 phone claims
// (0.314, 0.498). The clean attempt has no errors. Replace these only with a
// larger labelled set — the fixtures cannot measure false positives on their
// own.
const (
	MinWordConfidence  = 0.65
	MinPhoneConfidence = 0.70
	MaxShownErrors     = 3
)

// GateErrors drops low-confidence words, prunes low-confidence phones from
// the words that survive, then keeps the highest-confidence MaxShownErrors.
// The caller keeps the pre-gate length separately so the UI can say
// "showing 3 of 7" honestly.
func GateErrors(errors []MispronouncedWord) []MispronouncedWord {
	kept := make([]MispronouncedWord, 0, len(errors))
	for _, e := range errors {
		if e.Confidence < MinWordConfidence {
			continue
		}
		phones := make([]PhoneDetail, 0, len(e.Phones))
		for _, p := range e.Phones {
			if p.Confidence >= MinPhoneConfidence {
				phones = append(phones, p)
			}
		}
		e.Phones = phones
		kept = append(kept, e)
	}
	sort.SliceStable(kept, func(i, j int) bool {
		return kept[i].Confidence > kept[j].Confidence
	})
	if len(kept) > MaxShownErrors {
		kept = kept[:MaxShownErrors]
	}
	return kept
}
