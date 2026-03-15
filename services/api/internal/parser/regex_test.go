package parser

import (
	"testing"
)

func TestParseQuickAdd(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedAmt   *float64
		expectedDesc  *string
	}{
		{
			name:         "Complete phrase",
			input:        "Gasté 50 en café",
			expectedAmt:  ptrFloat(50.0),
			expectedDesc: ptrString("café"),
		},
		{
			name:         "Complete phrase with por",
			input:        "pagué 100.5 por gasolina",
			expectedAmt:  ptrFloat(100.5),
			expectedDesc: ptrString("gasolina"),
		},
		{
			name:         "Missing description",
			input:        "fueron 20",
			expectedAmt:  ptrFloat(20.0),
			expectedDesc: nil,
		},
		{
			name:         "Missing amount",
			input:        "compré de pan",
			expectedAmt:  nil,
			expectedDesc: ptrString("pan"),
		},
		{
			name:         "Missing keywords but has number and description",
			input:        "30 en dulces",
			expectedAmt:  ptrFloat(30.0),
			expectedDesc: ptrString("dulces"),
		},
		{
			name:         "Messy text",
			input:        "ehh gasté creo que 40,5 en comida rapida",
			expectedAmt:  ptrFloat(40.5),
			expectedDesc: ptrString("comida rapida"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := ParseQuickAdd(tc.input)
			
			// Compare Amount
			if tc.expectedAmt == nil && res.Amount != nil {
				t.Errorf("Expected nil amount, got %v", *res.Amount)
			} else if tc.expectedAmt != nil && res.Amount == nil {
				t.Errorf("Expected amount %v, got nil", *tc.expectedAmt)
			} else if tc.expectedAmt != nil && res.Amount != nil && *tc.expectedAmt != *res.Amount {
				t.Errorf("Expected amount %v, got %v", *tc.expectedAmt, *res.Amount)
			}

			// Compare Description
			if tc.expectedDesc == nil && res.Description != nil {
				t.Errorf("Expected nil description, got %v", *res.Description)
			} else if tc.expectedDesc != nil && res.Description == nil {
				t.Errorf("Expected description %v, got nil", *tc.expectedDesc)
			} else if tc.expectedDesc != nil && res.Description != nil && *tc.expectedDesc != *res.Description {
				t.Errorf("Expected description %v, got %v", *tc.expectedDesc, *res.Description)
			}
		})
	}
}

func ptrFloat(f float64) *float64 { return &f }
func ptrString(s string) *string { return &s }
