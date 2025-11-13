package components

// typeColor returns the Tailwind color name for a given message type.
func typeColor(t string) string {
	switch t {
	case "FPL":
		return "sky"
	case "CNL":
		return "rose"
	case "DEP":
		return "emerald"
	case "ARR":
		return "violet"
	case "DLA":
		return "amber"
	default:
		return "slate"
	}
}
