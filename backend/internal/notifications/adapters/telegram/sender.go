package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"qmanga/internal/notifications/application"
	"strings"
	"time"
)

type Sender struct {
	client                 *http.Client
	endpoint, chat, origin string
}

func New(token, chat, origin string) (*Sender, error) {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || strings.TrimSpace(token) == "" || strings.TrimSpace(chat) == "" {
		return nil, errors.New("invalid Telegram configuration")
	}
	return &Sender{client: &http.Client{Timeout: 12 * time.Second}, endpoint: "https://api.telegram.org/bot" + token + "/sendMessage", chat: chat, origin: strings.TrimRight(origin, "/")}, nil
}
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
func (s *Sender) Send(ctx context.Context, m application.ModerationMessage) error {
	labels := map[string]string{"team": "Заявка команды", "chapter": "Глава на проверку", "project": "Заявка на проект", "title_change": "Правки тайтла", "report": "Жалоба"}
	kind := labels[m.Kind]
	if kind == "" {
		kind = "Новая заявка"
	}
	if !strings.HasPrefix(m.Path, "/") || strings.HasPrefix(m.Path, "//") {
		return errors.New("invalid moderation destination")
	}
	text := "Qmanga · " + kind + "\n\n" + truncate(m.Label, 400) + "\n" + truncate(m.Detail, 1800) + "\n\nID: " + truncate(m.EntityID, 120)
	button := "Открыть заявку"
	if m.Kind == "team" || m.Kind == "project" || m.Kind == "report" {
		button = "Открыть очередь"
	}
	payload := map[string]any{"chat_id": s.chat, "text": text, "link_preview_options": map[string]bool{"is_disabled": true}, "reply_markup": map[string]any{"inline_keyboard": [][]map[string]string{{{"text": button, "url": s.origin + m.Path}}}}}
	body, err := json.Marshal(payload)
	if err != nil {
		return errors.New("Telegram encoding failed")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("Telegram request failed")
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return errors.New("Telegram transport failed")
	}
	defer response.Body.Close()
	var result struct {
		OK bool `json:"ok"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result) != nil || !result.OK {
		return errors.New("Telegram delivery rejected")
	}
	return nil
}
