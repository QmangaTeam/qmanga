package http

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

type PublishingHandler struct {
	service  *application.Service
	identity *IdentityHandler
	uploads  chan struct{}
	media    chan struct{}
}

func NewPublishingHandler(service *application.Service) *PublishingHandler {
	return &PublishingHandler{service: service, uploads: make(chan struct{}, 2), media: make(chan struct{}, 12)}
}
func (handler *PublishingHandler) Register(router *gin.Engine, identity *IdentityHandler) {
	handler.identity = identity
	api := router.Group("/api/v1")
	api.GET("/releases/:id/manifest", handler.manifest)
	api.GET("/releases/:id/pages/:page_id/original", handler.original)
	api.GET("/media/:asset_id", handler.asset)
	api.HEAD("/media/:asset_id", handler.asset)
	router.GET("/media/:asset_id", handler.asset)
	router.HEAD("/media/:asset_id", handler.asset)
	studio := api.Group("/studio", identity.RequireUser())
	studio.GET("/projects/:id/chapters", handler.projectChapters)
	studio.GET("/projects/:id/revisions", handler.projectRevisions)
	studio.POST("/projects/:id/revisions", handler.createDraft)
	studio.DELETE("/revisions/:id", handler.discardDraft)
	studio.PUT("/projects/:id/cover", handler.selectCover)
	studio.PUT("/projects/:id/cover-file", handler.uploadCover)
	studio.POST("/revisions/:id/pages/:page_id/retry", handler.retryPage)
	studio.GET("/revisions/:id", handler.revision)
	studio.PUT("/revisions/:id/pages", handler.upload)
	studio.PUT("/revisions/:id/order", handler.reorder)
	studio.DELETE("/revisions/:id/pages/:page_id", handler.deletePage)
	studio.POST("/revisions/:id/submit", handler.submit)
	studio.PUT("/revisions/:id/attribution", handler.attribution)
	studio.GET("/revisions/:id/manifest", handler.preview)
	moderation := api.Group("/moderation", identity.RequireUser(), RequireRole("admin", "moderator"))
	moderation.GET("/revisions", handler.queue)
	moderation.POST("/revisions/:id/review", handler.review)
	moderation.POST("/releases/:id/withdraw", handler.withdraw)
	moderation.POST("/releases/:id/restore", handler.restore)
}
func (handler *PublishingHandler) actor(request *gin.Context) string {
	if account, exists := CurrentAccount(request); exists {
		return account.ID
	}
	if handler.identity == nil || handler.identity.service == nil {
		return ""
	}
	raw, failure := request.Cookie("qmanga_session")
	if failure != nil {
		return ""
	}
	_, account, failure := handler.identity.service.Authenticate(request.Request.Context(), raw)
	if failure != nil {
		return ""
	}
	return account.ID
}
func revisionView(item application.StudioRevision) gin.H {
	pages := []gin.H{}
	for _, page := range item.Pages {
		pages = append(pages, gin.H{"id": page.ID, "position": page.Position, "width": page.Width, "height": page.Height, "bytes": page.Bytes, "status": page.Status, "failure": page.Failure, "upload_key": page.UploadKey, "original_name": page.OriginalName, "thumbnail_id": page.ThumbnailID})
	}
	return gin.H{"cover_id": item.CoverID, "release_status": item.ReleaseStatus, "preparation_method": item.PreparationMethod, "human_reviewed": item.HumanReviewed, "human_reviewer": item.HumanReviewer, "human_reviewed_at": item.HumanReviewedAt, "source_note": item.Source, "credits": item.Credits, "rights_confirmed": item.RightsConfirmed, "confirmed_by": item.ConfirmedBy, "id": item.ID, "release_id": item.ReleaseID, "project_id": item.ProjectID, "chapter_id": item.ChapterID, "title_id": item.TitleID, "title_name": item.TitleName, "chapter_label": item.ChapterLabel, "team_name": item.TeamName, "state": item.State, "note": item.Note, "version": item.Version, "pages": pages}
}
func (handler *PublishingHandler) projectChapters(request *gin.Context) {
	items, failure := handler.service.ProjectChapters(request.Request.Context(), handler.actor(request), request.Param("id"), requestOffset(request))
	if failure != nil {
		publicationError(request, failure)
		return
	}
	views := []gin.H{}
	for _, item := range items {
		views = append(views, gin.H{"id": item.ID, "label": item.Label, "name": item.Name, "order": item.Order})
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}
func (handler *PublishingHandler) projectRevisions(request *gin.Context) {
	items, failure := handler.service.ProjectRevisions(request.Request.Context(), handler.actor(request), request.Param("id"), requestOffset(request))
	if failure != nil {
		publicationError(request, failure)
		return
	}
	views := []gin.H{}
	for _, item := range items {
		views = append(views, revisionView(item))
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}
func (handler *PublishingHandler) createDraft(request *gin.Context) {
	var input struct {
		ChapterID string `json:"chapter_id"`
	}
	if !decode(request, &input) {
		return
	}
	item, failure := handler.service.CreateDraft(request.Request.Context(), handler.actor(request), request.Param("id"), input.ChapterID)
	if failure != nil {
		publicationError(request, failure)
		return
	}
	request.JSON(201, revisionView(item))
}
func (handler *PublishingHandler) discardDraft(request *gin.Context) {
	version, valid := queryInteger(request, "version", 1, 2147483647)
	if !valid {
		return
	}
	if failure := handler.service.DiscardDraft(request.Request.Context(), handler.actor(request), request.Param("id"), int64(version)); failure != nil {
		publicationError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *PublishingHandler) revision(request *gin.Context) {
	item, failure := handler.service.Revision(request.Request.Context(), handler.actor(request), request.Param("id"))
	if failure != nil {
		publicationError(request, failure)
		return
	}
	request.JSON(200, revisionView(item))
}
func (handler *PublishingHandler) upload(request *gin.Context) {
	select {
	case handler.uploads <- struct{}{}:
		defer func() { <-handler.uploads }()
	default:
		request.Header("Retry-After", "2")
		Error(request, 429, "busy", "Жүктеу кезегі толы. Қайта көріңіз.")
		return
	}
	if _, failure := handler.service.Revision(request.Request.Context(), handler.actor(request), request.Param("id")); failure != nil {
		publicationError(request, failure)
		return
	}
	position, valid := queryInteger(request, "position", 1, 300)
	if !valid {
		return
	}
	controller := http.NewResponseController(request.Writer)
	_ = controller.SetReadDeadline(time.Now().Add(10 * time.Minute))
	defer controller.SetReadDeadline(time.Time{})
	request.Request.Body = http.MaxBytesReader(request.Writer, request.Request.Body, 25<<20)
	file, failure := os.CreateTemp("", "qmanga-upload-")
	if failure != nil {
		publicationError(request, failure)
		return
	}
	defer func() { file.Close(); os.Remove(file.Name()) }()
	hash := sha256.New()
	size, failure := io.Copy(io.MultiWriter(file, hash), request.Request.Body)
	if failure != nil {
		Error(request, 413, "upload_size", "Бір сурет 25 MiB-тан аспауы керек")
		return
	}
	if _, failure = file.Seek(0, io.SeekStart); failure != nil {
		publicationError(request, failure)
		return
	}
	configuration, format, failure := image.DecodeConfig(file)
	if failure != nil {
		publicationError(request, domain.ErrInvalidImage)
		return
	}
	if _, failure = file.Seek(0, io.SeekStart); failure != nil {
		publicationError(request, failure)
		return
	}
	reservation, failure := handler.service.Upload(request.Request.Context(), handler.actor(request), application.Upload{OriginalName: request.Query("name"), RevisionID: request.Param("id"), Key: request.Query("key"), Position: position, Bytes: size, Width: configuration.Width, Height: configuration.Height, Format: format, Checksum: fmt.Sprintf("%x", hash.Sum(nil))}, file)
	if failure != nil {
		publicationError(request, failure)
		return
	}
	status := 201
	if reservation.Stored {
		status = 200
	}
	request.JSON(status, gin.H{"page_id": reservation.PageID, "stored": true})
}
func (handler *PublishingHandler) reorder(request *gin.Context) {
	var input struct {
		Version int64    `json:"version"`
		PageIDs []string `json:"page_ids"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.Reorder(request.Request.Context(), handler.actor(request), request.Param("id"), input.Version, input.PageIDs); failure != nil {
		publicationError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *PublishingHandler) deletePage(request *gin.Context) {
	version, valid := queryInteger(request, "version", 1, 2147483647)
	if !valid {
		return
	}
	if failure := handler.service.DeletePage(request.Request.Context(), handler.actor(request), request.Param("id"), request.Param("page_id"), int64(version)); failure != nil {
		publicationError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *PublishingHandler) submit(request *gin.Context) {
	var input struct {
		Version int64 `json:"version"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.Submit(request.Request.Context(), handler.actor(request), request.Param("id"), input.Version); failure != nil {
		publicationError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *PublishingHandler) queue(request *gin.Context) {
	items, failure := handler.service.FilteredModerationQueue(request.Request.Context(), handler.actor(request), request.Query("team"), request.Query("kind"), requestOffset(request))
	if failure != nil {
		publicationError(request, failure)
		return
	}
	views := []gin.H{}
	for _, item := range items {
		views = append(views, revisionView(item))
	}
	request.JSON(200, gin.H{"items": views, "next_cursor": nil})
}
func (handler *PublishingHandler) review(request *gin.Context) {
	var input struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.Review(request.Request.Context(), handler.actor(request), request.Param("id"), input.Decision, input.Reason); failure != nil {
		publicationError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *PublishingHandler) withdraw(request *gin.Context) {
	var input struct {
		Reason string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.Withdraw(request.Request.Context(), handler.actor(request), request.Param("id"), input.Reason); failure != nil {
		publicationError(request, failure)
		return
	}
	request.Status(204)
}
func publicationError(request *gin.Context, failure error) {
	switch {
	case errors.Is(failure, domain.ErrForbidden):
		Error(request, 403, "forbidden", "Бұл жобаға рұқсат жоқ")
	case errors.Is(failure, domain.ErrNotFound):
		Error(request, 404, "not_found", "Жарияланым табылмады")
	case errors.Is(failure, domain.ErrWithdrawn):
		Error(request, 410, "withdrawn", "Жарияланым оқудан алынды")
	case errors.Is(failure, domain.ErrConflict):
		Error(request, 409, "conflict", "Нұсқа өзгерді. Қайта жүктеңіз")
	case errors.Is(failure, domain.ErrImmutable), errors.Is(failure, domain.ErrTransition):
		Error(request, 409, "state", "Бет реті мен өңдеу күйін тексеріңіз")
	case errors.Is(failure, domain.ErrInvalidImage):
		Error(request, 422, "image", "JPEG, PNG немесе WebP таңдаңыз: 25 MiB, 80 MP; тарауға 300 бетке дейін")
	default:
		log.Printf("publication failure: %T", failure)
		Error(request, 503, "unavailable", "Қызмет уақытша қолжетімсіз")
	}
}
func (handler *PublishingHandler) attribution(request *gin.Context) {
	var input struct {
		Version           int64  `json:"version"`
		Source            string `json:"source_note"`
		Credits           string `json:"credits"`
		Confirmed         bool   `json:"rights_confirmed"`
		PreparationMethod string `json:"preparation_method"`
		HumanReviewed     bool   `json:"human_reviewed"`
	}
	if !decode(request, &input) {
		return
	}
	if failure := handler.service.SaveAttribution(request.Request.Context(), handler.actor(request), request.Param("id"), input.Version, domain.Attribution{Source: input.Source, Credits: input.Credits, Confirmed: input.Confirmed, PreparationMethod: input.PreparationMethod, HumanReviewed: input.HumanReviewed}); failure != nil {
		publicationError(request, failure)
		return
	}
	request.Status(204)
}

func (handler *PublishingHandler) restore(request *gin.Context) {
	var input struct {
		Reason string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	if err := handler.service.Restore(request.Request.Context(), handler.actor(request), request.Param("id"), input.Reason); err != nil {
		publicationError(request, err)
		return
	}
	request.Status(204)
}

// Legacy manual artwork commands must not bypass the title-change review workflow.
func (handler *PublishingHandler) selectCover(request *gin.Context) {
	Error(request, 409, "moderation_required", "Мұқаба өзгерісін жоба деректері арқылы модерацияға жіберіңіз")
}
func (handler *PublishingHandler) uploadCover(request *gin.Context) {
	Error(request, 409, "moderation_required", "Мұқаба өзгерісін жоба деректері арқылы модерацияға жіберіңіз")
}
func (handler *PublishingHandler) retryPage(request *gin.Context) {
	if err := handler.service.RetryPage(request.Request.Context(), handler.actor(request), request.Param("id"), request.Param("page_id")); err != nil {
		publicationError(request, err)
		return
	}
	request.Status(204)
}
