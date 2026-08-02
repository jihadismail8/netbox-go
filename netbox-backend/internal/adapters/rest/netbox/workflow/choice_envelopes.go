package workflow

// NetBox represents model choices as {value,label} on reads while accepting
// the scalar value on writes. These labels are copied from the pinned English
// v4.4.6 baseline; the standalone server does not load Python or NetBox data.
var commonChoiceLabels = map[string]string{
	"planned":         "Planned",
	"staging":         "Staging",
	"active":          "Active",
	"decommissioning": "Decommissioning",
	"retired":         "Retired",
	"reserved":        "Reserved",
	"available":       "Available",
	"deprecated":      "Deprecated",
	"offline":         "Offline",
	"staged":          "Staged",
	"failed":          "Failed",
	"inventory":       "Inventory",
	"container":       "Container",
	"dhcp":            "DHCP",
	"slaac":           "SLAAC",
}

var airflowLabels = map[string]string{
	"front-to-rear": "Front to rear",
	"rear-to-front": "Rear to front",
	"left-to-right": "Left to right",
	"right-to-left": "Right to left",
	"side-to-rear":  "Side to rear",
	"rear-to-side":  "Rear to side",
	"bottom-to-top": "Bottom to top",
	"top-to-bottom": "Top to bottom",
	"passive":       "Passive",
	"mixed":         "Mixed",
}

func choiceEnvelope(value any, labels map[string]string) any {
	// NetBox ChoiceField returns null for a nullable/blank choice.
	if value == nil || value == "" {
		return nil
	}
	key, _ := value.(string)
	label, exists := labels[key]
	if !exists {
		// The pinned serializer preserves unknown database values and emits an
		// empty label. Domain validation prevents new unknown values, but this
		// keeps reads compatible if a database was edited out of band.
		label = ""
	}
	return map[string]any{"value": value, "label": label}
}
