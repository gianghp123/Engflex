// Package prompts builds the word-draft LLM prompt. Pure functions only: no
// config, no network, no database.
package prompts

import "strings"

// BuildWordDraftPrompt renders the dictionary brief for one term. `context`
// is the sentence the learner met the word in (may be empty): it selects the
// sense and is excluded from the model's own contexts — the app keeps it as
// the entry's first quote.
//
// `language` is the English name of the language the explanation fields
// (definition, longDefinition, stressTip, commonSlip, etymology) are written
// in; "" and "English" both mean the default English brief, byte-identical to
// the original. Whatever the language, the term itself, ipa, syllables,
// contexts, collocations and wordForms stay English — they are the learning
// material, not the explanation.
//
// Static few-shot examples are tested live (writing-score pattern); template
// variables stay strictly for dynamic parts.
func BuildWordDraftPrompt(term, context, language string) string {
	learner := ""
	if context != "" {
		learner = "The learner met it in this sentence: " + context + "\n" +
			"Use their sentence ONLY to pick the right sense — never repeat it as one of your own contexts.\n"
	}
	if language == "" {
		language = "English"
	}
	localized := !strings.EqualFold(language, "English")
	intro := "You are a dictionary assistant for Vietnamese learners of English. Produce ONE dictionary entry " +
		"for the term below. Write everything in English.\n\n"
	if localized {
		intro = "You are a dictionary assistant for Vietnamese learners of English. Produce ONE dictionary entry " +
			"for the term below. Write the explanations in " + language + " — definition, longDefinition, stressTip, " +
			"commonSlip and etymology. Keep the rest English: the term itself, ipa, syllables, contexts, " +
			"collocations and wordForms.\n\n"
	}
	// Vietnamese gets its own calibrated example; any other language keeps the
	// English one with a note that it demonstrates structure only. Untested
	// territory: the probe covers en and vi.
	examples := wordDraftExamples
	if strings.EqualFold(language, "Vietnamese") {
		examples = wordDraftExamplesVietnamese
	} else if localized {
		examples = "The example below is in English only to show the JSON shape — write your explanation fields in " +
			language + ".\n\n" + wordDraftExamples
	}
	return intro +
		"Term: " + term + "\n" +
		learner + "\n" +
		"Rules:\n" +
		"1. definition is one learner-friendly " + language + " sentence; never a synonym list.\n" +
		"2. cefr is your best estimate of the term's level: exactly one of A1, A2, B1, B2, C1, C2.\n" +
		"3. partOfSpeech is the main part of speech of THIS sense; multi-word terms use phrase, idiom, or phrasal_verb.\n" +
		"4. ipa is the IPA transcription WITHOUT slashes; syllables counts every syllable, dot-separated " +
		"(the word punctual is punc.tu.al); stressTip tells the learner where the stress falls in one line.\n" +
		"5. commonSlip is the mistake Vietnamese learners commonly make with this word.\n" +
		"6. longDefinition is 2-3 sentences; contrast with near-synonyms when helpful.\n" +
		"7. etymology only when you are sure; otherwise empty string. NEVER invent an origin.\n" +
		"8. contexts holds 2-4 complete example sentences, each with a short label.\n" +
		"9. collocations holds 3-6 phrases that commonly go with the word, each with its pattern and one " +
		"complete example sentence.\n" +
		"10. wordForms holds up to 4 form relationships (plural, past tense, noun form, antonym...) with the " +
		"actual forms; [] when none apply.\n\n" +
		"Reply with a single JSON object and nothing else.\n\n" +
		examples
}

// wordDraftExamples is the static few-shot block (writing-score pattern): one
// full entry that calibrates structure and style, plus an explicit NOT list
// of the failure modes. Tested live on 2026-10-17; do not edit without
// re-running the probe (see the plan's Task 3 validation note).
const wordDraftExamples = `Examples:

Term: reliable
The learner met it in this sentence: She is the most reliable teammate I have ever worked with.

Output:
{"definition":"Able to be trusted to do what is expected, again and again.","cefr":"B1","partOfSpeech":"adjective","details":{"ipa":"rɪˈlaɪ.ə.bəl","syllables":"re.li.a.ble","stressTip":"Stress the second syllable: re-LI-a-ble.","commonSlip":"Learners stress the first syllable (RE-liable) or misspell it as 'relieable'.","longDefinition":"Reliable describes people or things you can depend on: they do what they promise, and they do not break down. It is stronger than 'good' — it is about trust over time. A reliable bus comes every morning; a reliable friend keeps their word.","etymology":"From 'rely' — to depend on someone.","contexts":[{"label":"At work","quote":"The whole team was impressed by how reliable she is under pressure."},{"label":"Everyday","quote":"My old bike is slow, but it is completely reliable."}],"collocations":[{"phrase":"a reliable source","pattern":"adj + noun","example":"According to a reliable source, the shop opens on Sunday."},{"phrase":"completely reliable","pattern":"adv + adj","example":"The wifi here is completely reliable."}],"wordForms":[{"formType":"noun","forms":["reliability"]},{"formType":"adverb","forms":["reliably"]},{"formType":"antonym","forms":["unreliable"]}]}}

---

NOT (these are the failure modes to avoid):
- Never answer the definition with a synonym list ("trustworthy, dependable") — always a full sentence.
- Never invent an etymology — empty string when unsure.
- Every context and collocation example must be a complete sentence.
- contexts/collocations/wordForms are arrays — [] is correct, never a missing key.
`

// wordDraftExamplesVietnamese calibrates the localized brief: the explanation
// fields in Vietnamese, every example sentence still English — the mix the
// learner sees on the word card. Tested live on 2026-10-17; do not edit
// without re-running the probe.
const wordDraftExamplesVietnamese = `Examples:

Term: reliable
The learner met it in this sentence: She is the most reliable teammate I have ever worked with.

Output:
{"definition":"Đáng tin cậy, có thể trông cậy được vì luôn làm đúng điều đã hứa.","cefr":"B1","partOfSpeech":"adjective","details":{"ipa":"rɪˈlaɪ.ə.bəl","syllables":"re.li.a.ble","stressTip":"Nhấn trọng âm ở âm tiết thứ hai: re-LI-a-ble.","commonSlip":"Người học Việt Nam hay nhấn sai âm tiết đầu (RE-liable) hoặc viết nhầm thành 'relieable'.","longDefinition":"Reliable miêu tả người hay vật mà bạn có thể trông cậy: họ làm đúng điều đã hứa và không bỏ dở việc. Từ này mạnh hơn 'good' — nó nói về sự tin cậy lâu dài. Một chuyến xe buýt reliable đến đều mỗi sáng; một người bạn reliable luôn giữ đúng lời.","etymology":"Xuất phát từ 'rely' — nghĩa là dựa vào, trông cậy vào ai đó.","contexts":[{"label":"At work","quote":"The whole team was impressed by how reliable she is under pressure."},{"label":"Everyday","quote":"My old bike is slow, but it is completely reliable."}],"collocations":[{"phrase":"a reliable source","pattern":"adj + noun","example":"According to a reliable source, the shop opens on Sunday."},{"phrase":"completely reliable","pattern":"adv + adj","example":"The wifi here is completely reliable."}],"wordForms":[{"formType":"noun","forms":["reliability"]},{"formType":"adverb","forms":["reliably"]},{"formType":"antonym","forms":["unreliable"]}]}}

---

NOT (these are the failure modes to avoid):
- Never answer the definition with a synonym list ("trustworthy, dependable") — always a full sentence in Vietnamese.
- Never invent an etymology — empty string when unsure.
- Explanation fields (definition, longDefinition, stressTip, commonSlip, etymology) are Vietnamese; the term, ipa, syllables, contexts, collocations and wordForms stay English.
- Every context and collocation example must be a complete English sentence.
- contexts/collocations/wordForms are arrays — [] is correct, never a missing key.
`
