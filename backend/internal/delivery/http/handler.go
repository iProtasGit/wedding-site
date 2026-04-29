package http

import (
	"wedding-app/internal/domain"
	"wedding-app/internal/telegram"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type RSVPHandler struct {
	useCase domain.RSVPUseCase
	tgBot   *telegram.Bot
}

func NewRSVPHandler(useCase domain.RSVPUseCase) *RSVPHandler {
	return &RSVPHandler{useCase: useCase}
}

func (h *RSVPHandler) HandleRSVP(c *fiber.Ctx) error {
	var req domain.RSVPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных. Пожалуйста, проверьте форму.",
		})
	}

	if err := h.useCase.SubmitRSVP(c.Context(), &req); err != nil {

		if h.tgBot != nil {
			go func() {
				// Context might be canceled, but we just want to fire and forget
				_ = h.tgBot.SendError(err)
			}()
		}
		// Differentiate between validation errors and server errors based on error message or type
		// For simplicity, we pass the usecase error directly if it's a validation error
		if err.Error() == "Список гостей пуст." ||
			err.Error() == "Имя гостя не может быть пустым." ||
			err.Error() == "Слишком длинное имя для гостя." ||
			len(err.Error()) > 40 && err.Error()[:40] == "Пожалуйста, уточните ваши предпочтения по" {

			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		log.Errorf("Error: %v", err)
		// Generic server error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Не удалось сохранить данные. Пожалуйста, попробуйте еще раз.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "success",
	})
}
