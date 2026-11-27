package ielts

import "fmt"

func validateMedia(content *Content) error {
	assets := make(map[string]Asset, len(content.Assets))
	for _, asset := range content.Assets {
		if asset.SourceID == "" {
			return fmt.Errorf("IELTS asset has empty sourceId")
		}
		if _, exists := assets[asset.SourceID]; exists {
			return fmt.Errorf("duplicate IELTS asset: %s", asset.SourceID)
		}
		assets[asset.SourceID] = asset
	}
	for _, section := range content.Sections {
		if section.AudioAssetID != nil {
			asset, ok := assets[*section.AudioAssetID]
			if !ok || asset.Type != "audio" {
				return fmt.Errorf("section %d: unresolved audio asset", section.SectionNumber)
			}
		}
		for _, group := range section.QuestionGroups {
			if err := validateContext(group.Context, assets, 0); err != nil {
				return fmt.Errorf("group %s: %w", group.ID, err)
			}
		}
	}
	return nil
}

func validateContext(blocks []ContextBlock, assets map[string]Asset, depth int) error {
	if depth > 32 {
		return fmt.Errorf("context nesting exceeds 32 levels")
	}
	for _, block := range blocks {
		switch block.Tag {
		case "text", "div", "p", "span", "b", "strong", "i", "em", "u", "br", "ul", "ol", "li",
			"table", "thead", "tbody", "tfoot", "tr", "td", "th", "img", "h1", "h2", "h3", "h4", "h5", "h6", "sup", "sub":
		default:
			return fmt.Errorf("unsupported context tag: %s", block.Tag)
		}
		if block.ColSpan < 0 || block.RowSpan < 0 {
			return fmt.Errorf("invalid table cell span")
		}
		if block.Tag == "img" && block.AssetID == nil {
			return fmt.Errorf("context image has no assetId")
		}
		if block.AssetID != nil {
			asset, ok := assets[*block.AssetID]
			if !ok || asset.Type != "image" || block.Tag != "img" {
				return fmt.Errorf("unresolved context image asset")
			}
		}
		if err := validateContext(block.Children, assets, depth+1); err != nil {
			return err
		}
	}
	return nil
}
