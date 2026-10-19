package http

import (
	"context"
	"crypto/subtle"
	"errors"
	"github.com/gin-gonic/gin"
	"qmanga/internal/identity/domain"
	"time"
)

type googleLinkService interface {
	GoogleLinked(context.Context, string) (bool, error)
	StartLink(context.Context, string, string) (string, string, error)
	FinishLink(context.Context, string, string, string) error
}

func (h *GoogleHandler) linkStatus(c *gin.Context) {
	service, ok := h.service.(googleLinkService)
	if !ok {
		c.JSON(200, gin.H{"enabled": false, "linked": false})
		return
	}
	account, _ := CurrentAccount(c)
	linked, err := service.GoogleLinked(c.Request.Context(), account.ID)
	if err != nil {
		h.identity.respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"enabled": true, "linked": linked})
}
func (h *GoogleHandler) linkStart(c *gin.Context) {
	account, _ := CurrentAccount(c)
	if !account.Verified {
		h.identity.respondError(c, domain.ErrForbidden)
		return
	}
	value, _ := c.Get("principal")
	principal := value.(principal)
	service, ok := h.service.(googleLinkService)
	if !ok {
		h.identity.respondError(c, domain.ErrNotFound)
		return
	}
	state, destination, err := service.StartLink(c.Request.Context(), account.ID, principal.session.Hash)
	if err != nil {
		h.identity.respondError(c, err)
		return
	}
	h.identity.cookie(c, "qmanga_google_state", state, time.Now().Add(10*time.Minute))
	h.identity.cookie(c, "qmanga_google_link", state, time.Now().Add(10*time.Minute))
	c.Header("Cache-Control", "private, no-store")
	c.JSON(200, gin.H{"url": destination})
}
func (h *GoogleHandler) linkCallback(c *gin.Context, state, code string) {
	service, ok := h.service.(googleLinkService)
	if !ok {
		c.Redirect(303, "/settings?google_link=error")
		return
	}
	raw, _ := c.Cookie("qmanga_session")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	err := service.FinishLink(ctx, state, code, raw)
	status := "success"
	if err != nil {
		status = "error"
		if errors.Is(err, domain.ErrConflict) {
			status = "conflict"
		}
	}
	c.Redirect(303, "/settings?google_link="+status)
}
func linkMarkerMatches(marker, state string) bool {
	return len(marker) == 43 && subtle.ConstantTimeCompare([]byte(marker), []byte(state)) == 1
}
