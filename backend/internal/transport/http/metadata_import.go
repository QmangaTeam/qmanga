package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"qmanga/internal/catalog/application"
)

type MetadataImportsHandler struct {
	service    *application.MetadataImports
	covers     chan struct{}
	coverQueue chan struct{}
}

func NewMetadataImportsHandler(service *application.MetadataImports) *MetadataImportsHandler {
	return &MetadataImportsHandler{service: service, covers: make(chan struct{}, 2), coverQueue: make(chan struct{}, 12)}
}
func (h *MetadataImportsHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	api.GET("/metadata-imports/thumbnail", identity.RequireUser(), h.cover)
	private := api.Group("/metadata-imports", identity.RequireUser(), expectedAccount())
	private.GET("/search", h.search)
	private.GET("/cover", h.cover)
	private.POST("/select", h.selectTitle)
	private.GET("/candidates", RequireRole("admin"), h.candidates)
	private.POST("/candidates/resolve", RequireRole("admin"), h.resolve)
}
func importError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrSourceLimited):
		c.Header("Retry-After", "3")
		Error(c, 429, "source_limited", "Дереккөз сұрау шегіне жетті. Сәл кейін қайталаңыз.")
	case errors.Is(err, application.ErrSourceUnavailable):
		Error(c, 502, "source_unavailable", "Дереккөз уақытша қолжетімсіз. Қайталап көріңіз.")
	default:
		catalogError(c, err)
	}
}
func (h *MetadataImportsHandler) search(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	actor, _ := CurrentAccount(c)
	items, err := h.service.Search(c.Request.Context(), actor.ID, c.Query("project"), c.Query("source"), c.Query("q"))
	if err != nil {
		importError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (h *MetadataImportsHandler) selectTitle(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	var input struct {
		Project string `json:"project"`
		Source  string `json:"source"`
		ID      string `json:"id"`
	}
	if !decode(c, &input) {
		return
	}
	actor, _ := CurrentAccount(c)
	item, err := h.service.Select(c.Request.Context(), actor.ID, input.Project, input.Source, input.ID)
	if err != nil {
		importError(c, err)
		return
	}
	c.JSON(200, item)
}
func (h *MetadataImportsHandler) candidates(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	actor, _ := CurrentAccount(c)
	items, err := h.service.Candidates(c.Request.Context(), actor.ID)
	if err != nil {
		importError(c, err)
		return
	}
	c.JSON(200, gin.H{"items": items})
}
func (h *MetadataImportsHandler) resolve(c *gin.Context) {
	var input struct {
		Source string `json:"source"`
		ID     string `json:"id"`
		Kind   string `json:"kind"`
		Target string `json:"target"`
	}
	if !decode(c, &input) {
		return
	}
	actor, _ := CurrentAccount(c)
	if err := h.service.Resolve(c.Request.Context(), actor.ID, input.Source, input.ID, input.Kind, input.Target); err != nil {
		importError(c, err)
		return
	}
	c.Status(204)
}

func (h *MetadataImportsHandler) cover(c *gin.Context) {
	select {
	case h.coverQueue <- struct{}{}:
		defer func() { <-h.coverQueue }()
	default:
		c.Header("Retry-After", "3")
		Error(c, 429, "busy", "Суретті өңдеу кезегі толы. Қайта көріңіз.")
		return
	}
	select {
	case h.covers <- struct{}{}:
		defer func() { <-h.covers }()
	case <-c.Request.Context().Done():
		return
	}
	c.Header("Cache-Control", "private, no-store")
	actor, _ := CurrentAccount(c)
	image, err := h.service.Cover(c.Request.Context(), actor.ID, c.Query("project"), c.Query("source"), c.Query("id"))
	if err != nil {
		importError(c, err)
		return
	}
	defer image.Close()
	c.Header("Content-Type", image.Asset.Mime)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Status(200)
	io.Copy(c.Writer, image.Content)
}
