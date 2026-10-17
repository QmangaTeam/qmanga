package application

import "qmanga/internal/publishing/domain"

func ValidateVariants(page Page, variants []Variant) error {
	expected := map[string][]domain.Fragment{}
	for tier, width := range map[string]int{"eco": 720, "standard": 1080, "high": 1600} {
		parts, failure := domain.PlanFragments(page.Width, page.Height, width)
		if failure != nil {
			return failure
		}
		expected[tier] = parts
	}
	expected["original"] = []domain.Fragment{{Width: page.Width, Height: page.Height}}
	seen := map[string]map[int]bool{}
	ids := map[string]bool{}
	for _, variant := range variants {
		if variant.ID == "" || variant.ObjectKey == "" || ids[variant.ID] || variant.Bytes <= 0 || len(variant.Checksum) != 64 || (variant.Mime != "image/webp" && variant.Mime != "image/png" && variant.Mime != "image/jpeg") {
			return domain.ErrInvalidImage
		}
		ids[variant.ID] = true
		fragments, exists := expected[variant.Tier]
		if !exists {
			return domain.ErrInvalidImage
		}
		if seen[variant.Tier] == nil {
			seen[variant.Tier] = map[int]bool{}
		}
		matched := false
		for _, fragment := range fragments {
			if fragment.Top == variant.Top && fragment.Width == variant.Width && fragment.Height == variant.Height && !seen[variant.Tier][variant.Top] {
				matched = true
				seen[variant.Tier][variant.Top] = true
				break
			}
		}
		if !matched {
			return domain.ErrInvalidImage
		}
	}
	for tier, fragments := range expected {
		if len(seen[tier]) != len(fragments) {
			return domain.ErrInvalidImage
		}
	}
	return nil
}
