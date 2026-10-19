package http

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"net/url"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	"strings"
	"time"
)

type googleLoginService interface {
	Start(context.Context, string, string) (string, string, error)
	Finish(context.Context, string, string) (application.LoginResult, string, error)
}
type GoogleHandler struct {
	service  googleLoginService
	identity *IdentityHandler
}

func NewGoogleHandler(service *application.GoogleService, identity *IdentityHandler) *GoogleHandler {
	handler := &GoogleHandler{identity: identity}
	if service != nil {
		handler.service = service
	}
	return handler
}
func (h *GoogleHandler) Register(api *gin.RouterGroup) {
	api.GET("/auth/google/status", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(200, gin.H{"enabled": h.service != nil})
	})
	if h.service == nil {
		return
	}
	api.POST("/auth/google/start", h.identity.CheckOrigin(), h.identity.limit(), h.start)
	api.GET("/auth/google/callback", h.identity.limit(), h.callback)
}
func (h *GoogleHandler) start(c *gin.Context) {
	var input struct {
		ReturnPath string `json:"return_path"`
		Nickname   string `json:"nickname"`
	}
	if !decode(c, &input) {
		return
	}
	state, destination, err := h.service.Start(c.Request.Context(), input.ReturnPath, input.Nickname)
	if err != nil {
		h.identity.respondError(c, err)
		return
	}
	h.identity.cookie(c, "qmanga_google_state", state, time.Now().Add(10*time.Minute))
	c.Header("Cache-Control", "no-store")
	c.JSON(200, gin.H{"url": destination})
}
func (h *GoogleHandler) callback(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	state := c.Query("state")
	bound, err := c.Cookie("qmanga_google_state")
	h.identity.cookie(c, "qmanga_google_state", "", time.Now())
	if err != nil || len(state) != 43 || subtle.ConstantTimeCompare([]byte(state), []byte(bound)) != 1 {
		Error(c, 400, "validation", "Кіруді қайта бастаңыз")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	var result application.LoginResult
	var destination, receipt string
	if service, ok := h.service.(interface {
		FinishWithGuestImport(context.Context, string, string, string) (application.LoginResult, string, error)
	}); ok {
		receipt, err = h.identity.registrationImportToken(c)
		if err == nil {
			result, destination, err = service.FinishWithGuestImport(ctx, state, c.Query("code"), receipt)
		}
	} else {
		result, destination, err = h.service.Finish(ctx, state, c.Query("code"))
	}
	if err != nil {
		if c.Query("error") == "access_denied" {
			c.Redirect(303, application.SafeAuthReturn(destination))
			return
		}
		code := "credentials"
		page := "/login"
		if errors.Is(err, application.ErrNicknameRequired) {
			code = "nickname"
			page = "/register"
		}
		if errors.Is(err, domain.ErrConflict) {
			code = "conflict"
		}
		c.Redirect(303, page+"?google_error="+code+"&returnTo="+url.QueryEscape(application.SafeAuthReturn(destination)))
		return
	}
	if receipt != "" {
		h.identity.rememberRegistrationImport(c, receipt)
	}
	h.identity.cookie(c, "qmanga_session", result.SessionToken, result.ExpiresAt)
	h.identity.cookie(c, "qmanga_csrf", result.CSRFToken, result.ExpiresAt)
	if result.Account.DisplayName == "" {
		destination = "/settings"
	}
	proof := sha256.Sum256([]byte(state + "\x00" + result.SessionToken))
	h.identity.cookie(c, "qmanga_external_complete", hex.EncodeToString(proof[:]), time.Now().Add(time.Minute))
	c.Redirect(303, googleCompletionURL(destination, state))
}

func googleCompletionURL(destination, nonce string) string {
	destination = application.SafeAuthReturn(destination)
	parsed, _ := url.Parse(destination)
	if parsed != nil && (parsed.Path == "/auth/complete" || strings.HasPrefix(parsed.Path, "/auth/complete/")) {
		destination = "/"
	}
	return "/auth/complete?returnTo=" + url.QueryEscape(destination) + "&nonce=" + url.QueryEscape(nonce)
}
