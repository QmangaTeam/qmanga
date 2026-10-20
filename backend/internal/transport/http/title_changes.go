package http

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	"strings"
)

type TitleChangesHandler struct {
	service *application.TitleChanges
	uploads chan struct{}
}

func NewTitleChangesHandler(service *application.TitleChanges) *TitleChangesHandler {
	return &TitleChangesHandler{service: service, uploads: make(chan struct{}, 2)}
}
func (h *TitleChangesHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	private := api.Group("", identity.RequireUser(), expectedAccount())
	private.GET("/studio/projects/:id/title-change-context", h.context)
	private.GET("/studio/projects/:id/title-changes", h.projectList)
	private.POST("/studio/projects/:id/title-changes", h.submit)
	private.GET("/title-changes/:id", h.detail)
	staff := private.Group("/moderation/title-changes", RequireRole("admin", "moderator"))
	staff.GET("", h.queue)
	staff.POST("/:id/review", h.review)
}
func changeView(item application.TitleChange) gin.H {
	assets := []gin.H{}
	for _, a := range item.Assets {
		assets = append(assets, gin.H{"id": a.ID, "purpose": a.Purpose, "url": coverURL(a.ID), "width": a.Width, "height": a.Height, "position": a.Position})
	}
	return gin.H{"id": item.ID, "project_id": item.ProjectID, "title_id": item.TitleID, "title_name": item.TitleName, "team_name": item.TeamName, "requested_name": item.RequestedName, "status": item.Status, "reason": item.Reason, "resolution": item.Resolution, "reviewed_name": item.ReviewedName, "base_version": item.BaseVersion, "cover_action": item.CoverAction, "banner_action": item.BannerAction, "cover_asset_id": item.CoverAssetID, "banner_asset_id": item.BannerAssetID, "links": item.Links, "before": item.Before, "proposed": item.Proposed, "assets": assets, "created_at": item.CreatedAt, "reviewed_at": item.ReviewedAt}
}
func changeError(c *gin.Context, err error) {
	if err == domain.ErrConflict {
		Error(c, 409, "conflict", "Деректер өзгерді немесе қаралуда өтінім бар. Бетті жаңартыңыз.")
		return
	}
	catalogError(c, err)
}
func (h *TitleChangesHandler) context(c *gin.Context) {
	account, _ := CurrentAccount(c)
	item, err := h.service.Context(c.Request.Context(), account.ID, c.Param("id"))
	if err != nil {
		changeError(c, err)
		return
	}
	c.JSON(200, gin.H{"project_id": item.ProjectID, "title_id": item.TitleID, "title_name": item.TitleName, "team_name": item.TeamName, "slug": item.Slug, "snapshot": item.Snapshot, "base_version": item.Snapshot.Version()})
}
func (h *TitleChangesHandler) submit(c *gin.Context) {
	account, _ := CurrentAccount(c)
	if _, err := h.service.Context(c.Request.Context(), account.ID, c.Param("id")); err != nil {
		changeError(c, err)
		return
	}
	select {
	case h.uploads <- struct{}{}:
		defer func() { <-h.uploads }()
	default:
		c.Header("Retry-After", "2")
		Error(c, 429, "busy", "Жүктеу кезегі толы. Қайта көріңіз.")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<20)
	if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
		Error(c, 422, "validation", "Файлдар көлемін және пішімін тексеріңіз")
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	form := c.Request.MultipartForm
	if len(form.Value) != 1 || len(form.Value["proposal"]) != 1 || len(form.Value["proposal"][0]) > 64<<10 {
		changeError(c, domain.ErrInvalid)
		return
	}
	var input domain.TitleChangeInput
	decoder := json.NewDecoder(strings.NewReader(form.Value["proposal"][0]))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
		changeError(c, domain.ErrInvalid)
		return
	}
	sources := []application.ChangeImageSource{}
	closers := []io.Closer{}
	defer func() {
		for _, f := range closers {
			f.Close()
		}
	}()
	for purpose, files := range form.File {
		limit := int64(25 << 20)
		maxCount := 1
		if purpose == "evidence" {
			limit = 5 << 20
			maxCount = 5
		} else if purpose != "cover" && purpose != "banner" {
			changeError(c, domain.ErrInvalid)
			return
		}
		if len(files) > maxCount {
			changeError(c, domain.ErrInvalid)
			return
		}
		for _, header := range files {
			if header.Size > limit {
				changeError(c, domain.ErrInvalid)
				return
			}
			file, err := header.Open()
			if err != nil {
				changeError(c, domain.ErrInvalid)
				return
			}
			closers = append(closers, file)
			sources = append(sources, application.ChangeImageSource{Purpose: purpose, Reader: file})
		}
	}
	item, err := h.service.Submit(c.Request.Context(), account.ID, c.Param("id"), input, sources)
	if err != nil {
		changeError(c, err)
		return
	}
	c.JSON(201, changeView(item))
}
func (h *TitleChangesHandler) projectList(c *gin.Context) {
	account, _ := CurrentAccount(c)
	items, err := h.service.ProjectList(c.Request.Context(), account.ID, c.Param("id"), c.Query("after"))
	h.list(c, items, err)
}
func (h *TitleChangesHandler) queue(c *gin.Context) {
	account, _ := CurrentAccount(c)
	items, err := h.service.Queue(c.Request.Context(), account.ID, c.Query("after"), c.Query("team"), c.Query("kind"))
	h.list(c, items, err)
}
func (h *TitleChangesHandler) list(c *gin.Context, items []application.TitleChange, err error) {
	if err != nil {
		changeError(c, err)
		return
	}
	views := []gin.H{}
	for _, item := range items {
		views = append(views, changeView(item))
	}
	c.JSON(200, gin.H{"items": views, "has_more": len(items) == 50})
}
func (h *TitleChangesHandler) detail(c *gin.Context) {
	account, _ := CurrentAccount(c)
	item, err := h.service.Detail(c.Request.Context(), account.ID, c.Param("id"))
	if err != nil {
		changeError(c, err)
		return
	}
	c.JSON(200, gin.H{"request": changeView(item.Request), "current": item.Current, "current_version": item.Current.Version(), "can_review": item.CanReview})
}
func (h *TitleChangesHandler) review(c *gin.Context) {
	var input struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if !decode(c, &input) {
		return
	}
	account, _ := CurrentAccount(c)
	if err := h.service.Review(c.Request.Context(), account.ID, c.Param("id"), input.Decision, input.Reason); err != nil {
		changeError(c, err)
		return
	}
	c.Status(204)
}
