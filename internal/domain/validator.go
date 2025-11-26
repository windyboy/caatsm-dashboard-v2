package domain

// Validator provides additional validation rules for telegrams
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateType checks if the telegram type is valid
func (v *Validator) ValidateType(telegramType string) bool {
	validTypes := map[string]bool{
		"AFTN":  true,
		"SITA":  true,
		"ACARS": true,
		"CPDLC": true,
	}
	return validTypes[telegramType]
}

// ValidatePriority checks if the priority is valid
func (v *Validator) ValidatePriority(priority int) bool {
	return priority >= 1 && priority <= 3
}

// ValidateICAOCode checks if a string is a valid ICAO code (4 characters, uppercase)
func (v *Validator) ValidateICAOCode(code string) bool {
	if len(code) != 4 {
		return false
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
