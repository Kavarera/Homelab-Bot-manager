package bot_test

import (
	"hs1-bot/internal/bot"
	"hs1-bot/internal/bot/middleware"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mockSender struct {
	sentMessages []tgbotapi.Chattable
	requests     []tgbotapi.Chattable
}

func (m *mockSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)
	return tgbotapi.Message{}, nil
}

func (m *mockSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.requests = append(m.requests, c)
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func TestRouter_CommandDispatch(t *testing.T) {
	sender := &mockSender{}
	router := bot.NewRouter(sender)

	var statusExecuted bool
	router.RegisterCommand("status", func(ctx *bot.Context) error {
		statusExecuted = true
		return ctx.Reply("status ok")
	})

	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 10,
			Chat:      &tgbotapi.Chat{ID: 12345},
			From:      &tgbotapi.User{ID: 999, UserName: "testuser"},
			Text:      "/status",
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 7},
			},
		},
	}

	router.HandleUpdate(update)

	if !statusExecuted {
		t.Error("expected /status command handler to be executed")
	}

	if len(sender.sentMessages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(sender.sentMessages))
	}
}

func TestRouter_CallbackDispatch(t *testing.T) {
	sender := &mockSender{}
	router := bot.NewRouter(sender)

	var callbackExecuted bool
	var selectedID string

	router.RegisterCallbackPrefix("action_select:", func(ctx *bot.Context) error {
		callbackExecuted = true
		selectedID = ctx.CallbackData
		return ctx.AnswerCallback("ok")
	})

	update := tgbotapi.Update{
		UpdateID: 2,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb123",
			From: &tgbotapi.User{ID: 999, UserName: "testuser"},
			Message: &tgbotapi.Message{
				MessageID: 20,
				Chat:      &tgbotapi.Chat{ID: 12345},
			},
			Data: "action_select:target-abc",
		},
	}

	router.HandleUpdate(update)

	if !callbackExecuted {
		t.Error("expected callback handler to be executed")
	}
	if selectedID != "action_select:target-abc" {
		t.Errorf("expected callback data 'action_select:target-abc', got '%s'", selectedID)
	}
}

func TestRouter_AuthMiddleware(t *testing.T) {
	sender := &mockSender{}
	router := bot.NewRouter(sender)

	const authorizedUserID int64 = 11111

	router.Use(middleware.Auth(authorizedUserID))

	var handled bool
	router.RegisterCommand("status", func(ctx *bot.Context) error {
		handled = true
		return nil
	})

	// 1. Unauthorized attempt
	unauthUpdate := tgbotapi.Update{
		UpdateID: 3,
		Message: &tgbotapi.Message{
			MessageID: 30,
			Chat:      &tgbotapi.Chat{ID: 22222},
			From:      &tgbotapi.User{ID: 22222, UserName: "intruder"},
			Text:      "/status",
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 7},
			},
		},
	}

	router.HandleUpdate(unauthUpdate)

	if handled {
		t.Error("unauthorized user should not be able to execute handler")
	}

	// Should have sent warning alert to authorized user
	if len(sender.sentMessages) != 1 {
		t.Fatalf("expected 1 alert sent to owner, got %d", len(sender.sentMessages))
	}

	// 2. Authorized attempt
	authUpdate := tgbotapi.Update{
		UpdateID: 4,
		Message: &tgbotapi.Message{
			MessageID: 31,
			Chat:      &tgbotapi.Chat{ID: 11111},
			From:      &tgbotapi.User{ID: 11111, UserName: "owner"},
			Text:      "/status",
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 7},
			},
		},
	}

	router.HandleUpdate(authUpdate)

	if !handled {
		t.Error("authorized user should execute handler")
	}
}
