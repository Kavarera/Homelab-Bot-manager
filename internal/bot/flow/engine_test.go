package flow_test

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot"
	"hs1-bot/internal/bot/flow"
	"hs1-bot/internal/bot/session"
	"hs1-bot/internal/bot/ui"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mockSender struct {
	sentMessages []tgbotapi.Chattable
}

func (m *mockSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)
	return tgbotapi.Message{}, nil
}

func (m *mockSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func createTestBotContext(userID int64, text string) *bot.Context {
	sender := &mockSender{}
	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 10,
			Chat:      &tgbotapi.Chat{ID: userID},
			From:      &tgbotapi.User{ID: userID, UserName: "tester"},
			Text:      text,
		},
	}
	return bot.NewContext(update, sender)
}

func TestFlowEngine_MultiStepWithKeyboards(t *testing.T) {
	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)

	wizardFlow := flow.NewBuilder("test_wizard").
		InitialStep("ask_choice", func(ctx *flow.Context) (flow.Action, error) {
			// Step 1: Prompt with choices + auto cancel button
			_ = ctx.ReplyWithStepKeyboard("Pilih opsi:", "Opsi A", "Opsi B")
			return ctx.Next("receive_choice"), nil
		}).
		Step("receive_choice", func(ctx *flow.Context) (flow.Action, error) {
			choice := ctx.RawText
			if choice != "Opsi A" && choice != "Opsi B" {
				_ = ctx.ReplyWithStepKeyboard("Pilihan tidak valid, silakan pilih:", "Opsi A", "Opsi B")
				return ctx.Stay(), nil
			}
			ctx.Set("choice", choice)
			// Step 2: Prompt free text + only cancel button
			_ = ctx.ReplyWithCancelKeyboard("Masukkan catatan tambahan:")
			return ctx.Next("receive_notes"), nil
		}).
		Step("receive_notes", func(ctx *flow.Context) (flow.Action, error) {
			savedChoice, _ := ctx.GetString("choice")
			notes := ctx.RawText
			// Step 3: Complete and remove custom keyboard
			_ = ctx.ReplyAndRemoveKeyboard(fmt.Sprintf("Selesai! Pilihan: %s, Catatan: %s", savedChoice, notes))
			return ctx.Complete(), nil
		}).
		Build()

	engine.Register(wizardFlow)

	const userID int64 = 88888

	// Step 1: Start flow
	startCtx := createTestBotContext(userID, "/start_wizard")
	err := engine.StartFlow(startCtx, "test_wizard")
	if err != nil {
		t.Fatalf("failed to start flow: %v", err)
	}

	// Verify session state is on "receive_choice"
	sess, err := store.Get(context.Background(), userID)
	if err != nil || sess == nil {
		t.Fatal("expected active session after start")
	}
	if sess.CurrentStep != "receive_choice" {
		t.Fatalf("expected step 'receive_choice', got '%s'", sess.CurrentStep)
	}

	// Step 2: User sends "Opsi A"
	choiceCtx := createTestBotContext(userID, "Opsi A")
	handled, err := engine.HandleActiveFlow(choiceCtx)
	if err != nil || !handled {
		t.Fatalf("expected choice input to be handled by active flow, handled=%v, err=%v", handled, err)
	}

	sess, _ = store.Get(context.Background(), userID)
	if sess.CurrentStep != "receive_notes" {
		t.Fatalf("expected step 'receive_notes', got '%s'", sess.CurrentStep)
	}

	// Step 3: User sends "Catatan penting"
	notesCtx := createTestBotContext(userID, "Catatan penting")
	handled, err = engine.HandleActiveFlow(notesCtx)
	if err != nil || !handled {
		t.Fatalf("expected notes input to be handled, handled=%v, err=%v", handled, err)
	}

	// Session should be cleared after Complete
	sess, _ = store.Get(context.Background(), userID)
	if sess != nil {
		t.Fatal("expected session to be cleared after completion")
	}
}

func TestFlowEngine_UserCancelButton(t *testing.T) {
	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)

	var cancelCalled bool
	testFlow := flow.NewBuilder("cancellable_flow").
		InitialStep("init", func(ctx *flow.Context) (flow.Action, error) {
			_ = ctx.ReplyWithCancelKeyboard("Ketik sesuatu:")
			return ctx.Next("step_1"), nil
		}).
		Step("step_1", func(ctx *flow.Context) (flow.Action, error) {
			return ctx.Stay(), nil
		}).
		OnCancel(func(ctx *flow.Context) error {
			cancelCalled = true
			return ctx.ReplyAndRemoveKeyboard("Dibatalkan!")
		}).
		Build()

	engine.Register(testFlow)

	const userID int64 = 77777
	startCtx := createTestBotContext(userID, "/start")
	_ = engine.StartFlow(startCtx, "cancellable_flow")

	// User clicks the "❌ Batal" button
	cancelCtx := createTestBotContext(userID, ui.DefaultCancelButton)
	handled, err := engine.HandleActiveFlow(cancelCtx)
	if err != nil || !handled {
		t.Fatalf("expected cancel button to be handled, handled=%v, err=%v", handled, err)
	}

	if !cancelCalled {
		t.Error("expected OnCancel handler to be executed when clicking cancel button")
	}

	sess, _ := store.Get(context.Background(), userID)
	if sess != nil {
		t.Error("expected session to be deleted after cancel")
	}
}
