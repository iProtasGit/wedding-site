package telegram

import (
	"fmt"
	"strings"
	"time"
	"wedding-app/internal/domain"

	"github.com/go-resty/resty/v2"
)

type Bot struct {
	token            string
	chatID           string
	countdownChatID  string
	countdownTopicID string
	client           *resty.Client
}

func NewBot(token, chatID, countdownChatID, countdownTopicID string) *Bot {
	if token == "" {
		return nil // Return nil if token is missing so we skip alerts/cron
	}
	return &Bot{
		token:            token,
		chatID:           chatID,
		countdownChatID:  countdownChatID,
		countdownTopicID: countdownTopicID,
		client:           resty.New(),
	}
}

func (b *Bot) SendAlert(req *domain.RSVPRequest) error {
	if b == nil || b.chatID == "" {
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

	return b.sendMessage(b.chatID, "", sb.String())
}

func (b *Bot) SendError(err error) error {
	if b == nil || b.chatID == "" {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("🔔 <b>ОШИБКА!</b>\n\n")
	sb.WriteString(err.Error())

	return b.sendMessage(b.chatID, "", sb.String())
}

func (b *Bot) SendCountdown(weddingDateStr string) error {
	if b == nil || b.countdownChatID == "" || weddingDateStr == "" {
		return nil
	}

	// Parse the wedding date
	weddingDate, err := time.Parse("2006-01-02T15:04:05", weddingDateStr)
	if err != nil {
		return fmt.Errorf("invalid wedding date format: %v", err)
	}

	now := time.Now()
	// Set both to midnight to count full days properly
	nowMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	// Apply UTC offset if needed, assuming Moscow Time (UTC+3) for logic simplicity or just use UTC
	nowMidnight = nowMidnight.Add(3 * time.Hour)

	weddingMidnight := time.Date(weddingDate.Year(), weddingDate.Month(), weddingDate.Day(), 0, 0, 0, 0, time.UTC)
	weddingMidnight = weddingMidnight.Add(3 * time.Hour)

	daysLeft := int(weddingMidnight.Sub(nowMidnight).Hours() / 24)

	var message string
	if daysLeft > 0 {
		message = fmt.Sprintf("💍 Доброе утро! До нашей свадьбы осталось: <b>%d %s</b>!", daysLeft, formatDays(daysLeft))
	} else if daysLeft == 0 {
		message = "🎉 Ура! Наша свадьба уже СЕГОДНЯ! Ждем вас с нетерпением!"
	} else {
		// Stop sending if the wedding passed
		return nil
	}

	return b.sendMessage(b.countdownChatID, b.countdownTopicID, message)
}

func (b *Bot) sendMessage(chatID, topicID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	body := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	if topicID != "" {
		body["message_thread_id"] = topicID
	}

	resp, err := b.client.R().
		SetBody(body).
		Post(url)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("telegram API error: %s", resp.String())
	}

	return nil
}

func formatDays(n int) string {
	if n%10 == 1 && n%100 != 11 {
		return "день"
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return "дня"
	}
	return "дней"
}
