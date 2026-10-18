package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"qmanga/internal/identity/application"
	"strconv"
	"strings"
	"time"
)

type AvatarHandler struct {
	service  *application.AvatarService
	identity *IdentityHandler
	slots    chan struct{}
}

func NewAvatarHandler(service *application.AvatarService, identity *IdentityHandler) *AvatarHandler {
	return &AvatarHandler{service, identity, make(chan struct{}, 4)}
}
func (h *AvatarHandler) Register(router *gin.Engine) {
	router.POST("/api/v1/avatars/user", h.identity.RequireUser(), h.upload("user"))
	router.POST("/api/v1/avatars/team/:id", h.identity.RequireUser(), h.upload("team"))
	router.GET("/media/avatars/:kind/:id", h.media)
	router.HEAD("/media/avatars/:kind/:id", h.media)
}
func (h *AvatarHandler) slot(c *gin.Context) bool {
	select {
	case h.slots <- struct{}{}:
		return true
	default:
		Error(c, 429, "rate_limit", "Сәл күтіңіз")
		return false
	}
}
func (h *AvatarHandler) upload(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.slot(c) {
			return
		}
		defer func() { <-h.slots }()
		mime := strings.Split(c.GetHeader("Content-Type"), ";")[0]
		if mime != "image/jpeg" && mime != "image/png" && mime != "image/webp" {
			Error(c, 400, "validation", "JPEG, PNG немесе WebP таңдаңыз")
			return
		}
		actor, _ := CurrentAccount(c)
		owner := actor.ID
		if kind == "team" {
			owner = c.Param("id")
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, application.MaxAvatarBytes)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		defer cancel()
		_, err := h.service.Upload(ctx, actor.ID, kind, owner, c.Request.Body)
		if err != nil {
			h.identity.respondError(c, err)
			return
		}
		c.JSON(200, gin.H{"url": "/media/avatars/" + kind + "/" + owner})
	}
}
func (h *AvatarHandler) media(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if !h.slot(c) {
		return
	}
	defer func() { <-h.slots }()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	avatar, stream, err := h.service.Open(ctx, c.Param("kind"), c.Param("id"))
	if err != nil {
		h.identity.respondError(c, err)
		return
	}
	defer stream.Close()
	etag := `"` + strings.TrimPrefix(avatar.Key, "avatars/") + `"`
	c.Header("ETag", etag)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(304)
		return
	}
	offset, length, status := int64(0), avatar.Bytes, 200
	if value := c.GetHeader("Range"); value != "" && c.Request.Method == "GET" {
		var valid bool
		offset, length, valid = mediaRange(value, avatar.Bytes)
		if !valid {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", avatar.Bytes))
			c.Status(416)
			return
		}
		status = 206
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, offset+length-1, avatar.Bytes))
	}
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", "image/jpeg")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Length", strconv.FormatInt(length, 10))
	c.Header("Last-Modified", avatar.UpdatedAt.Format(http.TimeFormat))
	if c.Request.Method == "HEAD" {
		c.Status(200)
		return
	}
	if offset > 0 {
		if _, err := io.CopyN(io.Discard, stream, offset); err != nil {
			c.Status(503)
			return
		}
	}
	c.Status(status)
	_, _ = io.CopyN(c.Writer, stream, length)
}
