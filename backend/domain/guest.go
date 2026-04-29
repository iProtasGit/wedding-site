package domain

// Guest represents an individual guest attending the wedding.
type Guest struct {
	FullName         string   `json:"fullName"`
	Alcohol          []string `json:"alcohol"`
	OtherAlcohol     string   `json:"otherAlcohol,omitempty"` // For "Other" alcohol preference
	TransferRequired bool     `json:"transferRequired"`
}

// RSVPRequest represents the complete RSVP submission from a user,
// potentially including multiple guests.
type RSVPRequest struct {
	Guests []Guest `json:"guests"`
}
