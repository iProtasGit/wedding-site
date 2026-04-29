package repository

import (
	"context"
	"fmt"
	"os"
	"strings"

	"wedding-app/internal/domain"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type sheetsRepo struct {
	srv           *sheets.Service
	spreadsheetID string
}

func NewSheetsRepository(credentialsFile, spreadsheetID string) (domain.RSVPRepository, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		fmt.Printf("Warning: unable to read client secret file %s: %v. Running in mock mode.\n", credentialsFile, err)
		return &sheetsRepo{spreadsheetID: spreadsheetID}, nil
	}

	// Important fix for "invalid_grant" / "Token must be a short-lived token"
	// We use google.CredentialsFromJSON instead of google.JWTConfigFromJSON directly
	// This automatically handles token refreshes and time syncing much better for Service Accounts
	creds, err := google.CredentialsFromJSON(ctx, b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to credentials: %v", err)
	}

	srv, err := sheets.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	return &sheetsRepo{srv: srv, spreadsheetID: spreadsheetID}, nil
}

func (r *sheetsRepo) SaveRSVP(ctx context.Context, req *domain.RSVPRequest) error {
	if r.srv == nil {
		fmt.Println("Mock SaveRSVP called (no credentials)")
		for _, g := range req.Guests {
			fmt.Printf("Mock Save: %+v\n", g)
		}
		return nil
	}

	var values [][]interface{}
	for _, guest := range req.Guests {
		alcoholStr := strings.Join(guest.Alcohol, ", ")
		if len(guest.Alcohol) == 0 {
			alcoholStr = "Не указано"
		}
		if guest.OtherAlcohol != "" {
			alcoholStr += fmt.Sprintf(" (Уточнение: %s)", guest.OtherAlcohol)
		}

		transferStr := "Нет"
		if guest.Transfer {
			transferStr = "Да"
		}

		values = append(values, []interface{}{guest.FullName, alcoholStr, transferStr})
	}

	var vr sheets.ValueRange
	vr.Values = values

	// Make sure to add the 'Transfer' column in Google Sheets!
	writeRange := "Sheet1!A:C"
	_, err := r.srv.Spreadsheets.Values.Append(r.spreadsheetID, writeRange, &vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("ошибка при записи в таблицу: %v", err)
	}

	return nil
}
