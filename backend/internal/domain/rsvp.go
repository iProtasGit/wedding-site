package domain

import "context"

type Guest struct {
	FullName     string   `json:"fullName"`
	Alcohol      []string `json:"alcohol"`
	OtherAlcohol string   `json:"otherAlcohol,omitempty"`
	Transfer     bool     `json:"transfer"`
}

type RSVPRequest struct {
	Guests []Guest `json:"guests"`
}

type RSVPRepository interface {
	SaveRSVP(ctx context.Context, req *RSVPRequest) error
}

type RSVPUseCase interface {
	SubmitRSVP(ctx context.Context, req *RSVPRequest) error
}
