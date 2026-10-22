package http

import (
	"github.com/gin-gonic/gin"
	"qmanga/internal/notifications/application"
)

type NotificationHandler struct{ service *application.Service }

func NewNotificationHandler(service *application.Service) *NotificationHandler {
	return &NotificationHandler{service}
}
func (handler *NotificationHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	group := api.Group("/notifications", identity.RequireUser(), expectedAccount())
	group.GET("", handler.list)
	group.GET("/unread-count", handler.unread)
	group.PUT("/read-all", handler.readAll)
	group.PUT("/:id/read", handler.read)
}
func (handler *NotificationHandler) list(request *gin.Context) {
	account, _ := CurrentAccount(request)
	kind := request.Query("kind")
	if kind != "" && kind != "chapters" && kind != "replies" {
		Error(request, 400, "invalid_kind", "Хабар санаты жарамсыз")
		return
	}
	items, failure := handler.service.ListFiltered(request.Request.Context(), account.ID, request.Query("after"), request.Query("unread") == "true", kind)
	if failure != nil {
		Error(request, 503, "unavailable", "Хабарлар уақытша қолжетімсіз")
		return
	}
	result := []gin.H{}
	for _, item := range items {
		result = append(result, gin.H{"id": item.ID, "kind": item.Kind, "title": item.Title, "body": item.Body, "release_id": item.ReleaseID, "title_slug": item.TitleSlug, "project_id": item.ProjectID, "comment_id": item.CommentID, "read": item.Read, "created_at": item.CreatedAt})
	}
	request.JSON(200, gin.H{"items": result, "has_more": len(items) == 50})
}
func (handler *NotificationHandler) read(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if failure := handler.service.Read(request.Request.Context(), account.ID, request.Param("id")); failure != nil {
		Error(request, 503, "unavailable", "Хабарды белгілеу мүмкін болмады")
		return
	}
	request.Status(204)
}

func (handler *NotificationHandler) unread(request *gin.Context) {
	account, _ := CurrentAccount(request)
	count, failure := handler.service.Unread(request.Request.Context(), account.ID)
	if failure != nil {
		Error(request, 503, "unavailable", "Хабарлар уақытша қолжетімсіз")
		return
	}
	request.JSON(200, gin.H{"count": count})
}

func (handler *NotificationHandler) readAll(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if err := handler.service.ReadAll(request.Request.Context(), account.ID); err != nil {
		Error(request, 503, "unavailable", "Хабарларды белгілеу мүмкін болмады")
		return
	}
	request.Status(204)
}
