package telegram

import (
	"fmt"
	"strings"
	"wedding-app/internal/domain"

	"github.com/go-resty/resty/v2"
)

type Bot struct {
	token  string
	chatID string
	client *resty.Client
}

func NewBot(token, chatID string) *Bot {
	if token == "" || chatID == "" {
		return nil // Return nil if config is missing so we can skip alerts
	}
	return &Bot{
		token:  token,
		chatID: chatID,
		client: resty.New(),
	}
}

func (b *Bot) SendAlert(req *domain.RSVPRequest) error {
	if b == nil {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("🔔 <b>Новая заявка RSVP!</b>\n\n")

	for i, guest := range req.Guests {
		sb.WriteString(fmt.Sprintf("<b>Гость %d:</b> %s\n", i+1, guest.FullName))

		alcohol := strings.Join(guest.Alcohol, ", ")
		if alcohol == "" {
			alcohol = "Не указано"
		}
		if guest.OtherAlcohol != "" {
			alcohol += fmt.Sprintf(" (Уточнение: %s)", guest.OtherAlcohol)
		}
		sb.WriteString(fmt.Sprintf("🍹 <b>Алкоголь:</b> %s\n", alcohol))

		transfer := "Нет"
		if guest.Transfer {
			transfer = "Да"
		}
		sb.WriteString(fmt.Sprintf("🚌 <b>Трансфер:</b> %s\n\n", transfer))
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	resp, err := b.client.R().
		SetBody(map[string]interface{}{
			"chat_id":    b.chatID,
			"text":       sb.String(),
			"parse_mode": "HTML",
		}).
		Post(url)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("telegram API error: %s", resp.String())
	}

	return nil
}

func (b *Bot) SendError(err error) error {
	if b == nil {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("🔔 <b>ОШИБКА!</b>\n\n")
	sb.WriteString(err.Error())

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	resp, err := b.client.R().
		SetBody(map[string]interface{}{
			"chat_id":    b.chatID,
			"text":       sb.String(),
			"parse_mode": "HTML",
		}).
		Post(url)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("telegram API error: %s", resp.String())
	}

	return nil
}
