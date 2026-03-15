package parser

import (
	"regexp"
	"strconv"
	"strings"
)

type ParsedTransaction struct {
	Amount      *float64
	Description *string
}

// ParseQuickAdd uses deterministic regex rules to find amounts and keywords.
func ParseQuickAdd(text string) ParsedTransaction {
	result := ParsedTransaction{}
	if text == "" {
		return result
	}

	// Clean up text
	text = strings.ToLower(strings.TrimSpace(text))

	// Regex to extract amount (digits optionally joined by dot or comma for cents)
	// Covers variations like "gasté 50", "pagué 15.5", "compré por 10"
	amountRegex := regexp.MustCompile(`(?:gast[eé]|pagu[eé]|compr[eé]|fueron|son)\s+(?:un|unos|una|unas)?\s*(\d+(?:[.,]\d+)?)`)
	
	amountMatches := amountRegex.FindStringSubmatch(text)
	if len(amountMatches) > 1 {
		// Normalize comma to dot for parsing
		amtStr := strings.ReplaceAll(amountMatches[1], ",", ".")
		if val, err := strconv.ParseFloat(amtStr, 64); err == nil {
			result.Amount = &val
		}
	} else {
		// Fallback: Just find the first number in the text if keywords are missing
		fallbackRegex := regexp.MustCompile(`(\d+(?:[.,]\d+)?)`)
		fallbackMatch := fallbackRegex.FindStringSubmatch(text)
		if len(fallbackMatch) > 1 {
			amtStr := strings.ReplaceAll(fallbackMatch[1], ",", ".")
			if val, err := strconv.ParseFloat(amtStr, 64); err == nil {
				result.Amount = &val
			}
		}
	}

	// Regex to extract description (what was bought)
	// Covers "en [algo]", "por [algo]", "de [algo]", supports accents
	descRegex := regexp.MustCompile(`(?:en|por|de)\s+([a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+)`)
	descMatches := descRegex.FindStringSubmatch(text)
	
	if len(descMatches) > 1 {
		// Clean up the match
		desc := strings.TrimSpace(descMatches[1])
		// Remove trailing words that might get caught if it's too broad, though [a-zA-Z\s] is gentle.
		if len(desc) > 0 {
			result.Description = &desc
		}
	}

	return result
}
