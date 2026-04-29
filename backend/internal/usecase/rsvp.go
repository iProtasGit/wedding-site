package usecase

import (
	"context"
	"errors"
	"wedding-app/internal/domain"
	"wedding-app/internal/telegram"
)

type rsvpUseCase struct {
	repo  domain.RSVPRepository
	tgBot *telegram.Bot
}

func NewRSVPUseCase(repo domain.RSVPRepository, tgBot *telegram.Bot) domain.RSVPUseCase {
	return &rsvpUseCase{
		repo:  repo,
		tgBot: tgBot,
	}
}

func (u *rsvpUseCase) SubmitRSVP(ctx context.Context, req *domain.RSVPRequest) error {
	if len(req.Guests) == 0 {
		return errors.New("Список гостей пуст.")
	}
	for i, g := range req.Guests {
		if g.FullName == "" {
			// To keep it simple, we just return a generic valid message. Handled better on frontend anyway.
			return errors.New("Имя гостя не может быть пустым.")
		}
		// Basic validation
		if len(g.FullName) > 100 {
			return errors.New("Слишком длинное имя для гостя.")
		}
		// If they chose "Другое", they must have specified what they want
		hasOther := false
		for _, a := range g.Alcohol {
			if a == "Другое" {
				hasOther = true
				break
			}
		}
		if hasOther && g.OtherAlcohol == "" {
			return errors.New("Пожалуйста, уточните ваши предпочтения по алкоголю для гостя №" + string(rune('1'+i)))
		}
	}

	err := u.repo.SaveRSVP(ctx, req)
	if err != nil {
		return err
	}

	// Try sending Telegram alert in background, do not block the request
	if u.tgBot != nil {
		go func() {
			// Context might be canceled, but we just want to fire and forget
			_ = u.tgBot.SendAlert(req)
		}()
	}

	return nil
}
