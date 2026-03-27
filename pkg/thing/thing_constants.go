package thing

// Defines values for ThingStatus.
const (
	Abandonné      ThingStatus = "Abandonné"
	Démoli         ThingStatus = "Démoli"
	EnConstruction ThingStatus = "En Construction"
	Planifié       ThingStatus = "Planifié"
	Utilisé        ThingStatus = "Utilisé"
)

// ThingStatus defines model for ThingStatus manually to preserve strongly typed strings.
type ThingStatus string
