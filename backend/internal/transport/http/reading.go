package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"qmanga/internal/reading/application"
	"qmanga/internal/reading/domain"
)

type ReadingHandler struct{ service *application.Service }

func NewReadingHandler(service *application.Service) *ReadingHandler { return &ReadingHandler{service} }
func (handler *ReadingHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	group := api.Group("", identity.RequireUser(), expectedAccount())
	group.GET("/library", handler.list)
	group.GET("/library/:id", handler.selection)
	group.PUT("/library/:id", handler.save)
	group.DELETE("/library/:id", handler.remove)
	group.GET("/progress/:id", handler.progress)
	group.POST("/progress/:id", handler.advance)
}
func expectedAccount() gin.HandlerFunc {
	return func(request *gin.Context) {
		account, _ := CurrentAccount(request)
		if request.GetHeader("X-Qmanga-Account") != account.ID {
			Error(request, 409, "account_changed", "Есептікжазба өзгерді. Бетті жаңартыңыз.")
			return
		}
		request.Header("Cache-Control", "private, no-store")
		request.Next()
	}
}
func progressView(item domain.Progress) gin.H {
	return gin.H{"title_id": item.TitleID, "release_id": item.ReleaseID, "revision_id": item.RevisionID, "page_id": item.PageID, "fraction": item.Fraction, "furthest_order": item.FurthestOrder, "version": item.Version, "device_id": item.DeviceID, "sequence": item.Sequence}
}
func entryView(item application.Entry) gin.H {
	return gin.H{"title_id": item.TitleID, "name": item.Name, "slug": item.Slug, "status": item.Status, "favorite": item.Favorite, "preferred_team_id": item.PreferredTeamID, "notify": item.Notify, "unseen": item.Unseen, "progress": progressView(item.Progress)}
}
func (handler *ReadingHandler) selection(request *gin.Context) {
	account, _ := CurrentAccount(request)
	item, failure := handler.service.Selection(request.Request.Context(), account.ID, request.Param("id"))
	if failure != nil {
		readingError(request, failure)
		return
	}
	if item == nil {
		request.JSON(200, gin.H{"entry": nil})
		return
	}
	request.JSON(200, gin.H{"entry": entryView(*item)})
}
func (handler *ReadingHandler) list(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.List(request.Request.Context(), account.ID, request.Query("after"))
	if failure != nil {
		readingError(request, failure)
		return
	}
	result := []gin.H{}
	for _, item := range items {
		result = append(result, gin.H{"title_id": item.TitleID, "name": item.Name, "slug": item.Slug, "status": item.Status, "favorite": item.Favorite, "preferred_team_id": item.PreferredTeamID, "notify": item.Notify, "unseen": item.Unseen, "progress": progressView(item.Progress)})
	}
	request.JSON(200, gin.H{"items": result, "has_more": len(items) == 100})
}
func (handler *ReadingHandler) save(request *gin.Context) {
	var input struct {
		Status          string `json:"status"`
		Favorite        bool   `json:"favorite"`
		PreferredTeamID string `json:"preferred_team_id"`
		Notify          bool   `json:"notify"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Save(request.Request.Context(), account.ID, application.Selection{TitleID: request.Param("id"), Status: input.Status, Favorite: input.Favorite, PreferredTeamID: input.PreferredTeamID, Notify: input.Notify}); failure != nil {
		readingError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *ReadingHandler) remove(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if failure := handler.service.Remove(request.Request.Context(), account.ID, request.Param("id")); failure != nil {
		readingError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *ReadingHandler) progress(request *gin.Context) {
	account, _ := CurrentAccount(request)
	item, failure := handler.service.Progress(request.Request.Context(), account.ID, request.Param("id"))
	if failure != nil {
		readingError(request, failure)
		return
	}
	request.JSON(200, progressView(item))
}
func (handler *ReadingHandler) advance(request *gin.Context) {
	var input struct {
		ID          string  `json:"event_id"`
		DeviceID    string  `json:"device_id"`
		Sequence    int64   `json:"sequence"`
		BaseVersion int64   `json:"base_version"`
		ReleaseID   string  `json:"release_id"`
		RevisionID  string  `json:"revision_id"`
		PageID      string  `json:"page_id"`
		Fraction    float64 `json:"fraction"`
		Completed   bool    `json:"completed"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	result, failure := handler.service.Advance(request.Request.Context(), account.ID, domain.Event{ID: input.ID, TitleID: request.Param("id"), DeviceID: input.DeviceID, Sequence: input.Sequence, BaseVersion: input.BaseVersion, ReleaseID: input.ReleaseID, RevisionID: input.RevisionID, PageID: input.PageID, Fraction: input.Fraction, Completed: input.Completed})
	if failure != nil {
		readingError(request, failure)
		return
	}
	status := 200
	if result.Outcome == "conflict" {
		status = 409
	}
	request.JSON(status, gin.H{"outcome": result.Outcome, "progress": progressView(result.Progress)})
}
func readingError(request *gin.Context, failure error) {
	switch {
	case errors.Is(failure, domain.ErrInvalid):
		Error(request, 422, "validation", "Оқу деректерін тексеріңіз")
	case errors.Is(failure, domain.ErrNotFound):
		Error(request, 404, "not_found", "Тарау немесе бет табылмады")
	case errors.Is(failure, domain.ErrForbidden):
		Error(request, 403, "forbidden", "Рұқсат жоқ")
	default:
		log.Printf("reading failure: %T", failure)
		Error(request, 503, "unavailable", "Оқу деректері уақытша қолжетімсіз")
	}
}
