package requests

import "encoding/json"

// JSON null clears nullable columns; omission preserves the previous value.
func (r *UpdateVocabularyItem) UnmarshalJSON(data []byte) error {
	type plain UpdateVocabularyItem
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*r = UpdateVocabularyItem(decoded)
	_, r.DomainSpecified = fields["domain"]
	_, r.PartOfSpeechSpecified = fields["partOfSpeech"]
	return nil
}
