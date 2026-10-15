package enums

// VocabularyPartOfSpeech is a word's part of speech
// (vocabulary_items.part_of_speech). Display labels are frontend copy.
type VocabularyPartOfSpeech string

const (
	VocabularyPartOfSpeechNoun         VocabularyPartOfSpeech = "noun"
	VocabularyPartOfSpeechVerb         VocabularyPartOfSpeech = "verb"
	VocabularyPartOfSpeechAdjective    VocabularyPartOfSpeech = "adjective"
	VocabularyPartOfSpeechAdverb       VocabularyPartOfSpeech = "adverb"
	VocabularyPartOfSpeechPronoun      VocabularyPartOfSpeech = "pronoun"
	VocabularyPartOfSpeechPreposition  VocabularyPartOfSpeech = "preposition"
	VocabularyPartOfSpeechConjunction  VocabularyPartOfSpeech = "conjunction"
	VocabularyPartOfSpeechInterjection VocabularyPartOfSpeech = "interjection"
	VocabularyPartOfSpeechDeterminer   VocabularyPartOfSpeech = "determiner"
	VocabularyPartOfSpeechPhrase       VocabularyPartOfSpeech = "phrase"
	VocabularyPartOfSpeechIdiom        VocabularyPartOfSpeech = "idiom"
	VocabularyPartOfSpeechPhrasalVerb  VocabularyPartOfSpeech = "phrasal_verb"
)
