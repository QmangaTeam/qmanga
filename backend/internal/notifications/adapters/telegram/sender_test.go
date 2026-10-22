package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"qmanga/internal/notifications/application"
	"strings"
	"testing"
)

func TestModerationMessage(t *testing.T) {
	var payload struct {
		Text        string `json:"text"`
		ChatID      string `json:"chat_id"`
		ReplyMarkup struct {
			InlineKeyboard [][]struct{ URL string } `json:"inline_keyboard"`
		} `json:"reply_markup"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("not POST")
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	sender, _ := New("secret-token", "123", "https://qmanga.example")
	sender.endpoint = server.URL
	err := sender.Send(context.Background(), application.ModerationMessage{Kind: "chapter", EntityID: "exact-id", Label: "Қазақша <title>", Detail: strings.Repeat("Ә", 6000), Path: "/admin/moderation?revision=exact-id"})
	if err != nil {
		t.Fatal(err)
	}
	if payload.ChatID != "123" || !strings.Contains(payload.Text, "Қазақша <title>") || len([]rune(payload.Text)) > 4096 || payload.ReplyMarkup.InlineKeyboard[0][0].URL != "https://qmanga.example/admin/moderation?revision=exact-id" {
		t.Fatal("unsafe or incorrect message")
	}
}
func TestRejectedAndTransportErrorsAreSanitized(t *testing.T) {
	for _, body := range []string{`{"ok":false}`, `not json`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
		sender, _ := New("secret-token", "123", "https://qmanga.example")
		sender.endpoint = server.URL
		err := sender.Send(context.Background(), application.ModerationMessage{Path: "/admin"})
		server.Close()
		if err == nil || strings.Contains(err.Error(), "secret-token") {
			t.Fatal("error not safe")
		}
	}
	sender, _ := New("secret-token", "123", "https://qmanga.example")
	sender.endpoint = "http://127.0.0.1:1/secret-token"
	if err := sender.Send(context.Background(), application.ModerationMessage{Path: "/admin"}); err == nil || strings.Contains(err.Error(), "secret-token") {
		t.Fatal("transport leaked token")
	}
	if err := sender.Send(context.Background(), application.ModerationMessage{Path: "//evil.example"}); err == nil {
		t.Fatal("external destination accepted")
	}
}
