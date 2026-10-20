package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"qmanga/internal/community/application"
	"qmanga/internal/community/domain"
)

type CommunityHandler struct{ service *application.Service }

func NewCommunityHandler(service *application.Service) *CommunityHandler {
	return &CommunityHandler{service}
}
func (handler *CommunityHandler) Register(api *gin.RouterGroup, identity *IdentityHandler) {
	api.GET("/community/titles/:id", handler.discussion)
	api.POST("/rights-reports", identity.CheckOrigin(), identity.limit(), handler.rightsReport)
	api.POST("/rights-reports/status", identity.CheckOrigin(), identity.limit(), handler.rightsStatus)
	api.GET("/community/comments/:id/replies", handler.replies)
	member := api.Group("/community", identity.RequireUser(), expectedAccount())
	member.GET("/titles/:id/rating", handler.myRating)
	member.PUT("/titles/:id/rating", handler.rating)
	member.POST("/comments", handler.comment)
	member.PUT("/comments/:id", handler.edit)
	member.DELETE("/comments/:id", handler.remove)
	member.POST("/reports", handler.report)
	member.GET("/reports", handler.myReports)
	moderator := api.Group("/moderation", identity.RequireUser(), expectedAccount(), RequireRole("admin", "moderator"))
	moderator.GET("/reports", handler.reports)
	moderator.PUT("/reports/:id", handler.resolve)
	moderator.PUT("/comments/:id/visibility", handler.visibility)
}
func (handler *CommunityHandler) discussion(request *gin.Context) {
	result, failure := handler.service.Discussion(request.Request.Context(), request.Param("id"), request.Query("release"), request.Query("after"), request.Query("comment"))
	if failure != nil {
		communityError(request, failure)
		return
	}
	items := []gin.H{}
	for _, item := range result.Comments {
		items = append(items, commentView(item))
	}
	request.JSON(200, gin.H{"comments": items, "next": result.Next, "score": result.Score, "ratings": result.Ratings})
}
func commentView(item application.Comment) gin.H {
	return gin.H{"id": item.ID, "user_id": item.UserID, "display_name": item.DisplayName, "parent_id": item.ParentID, "root_id": item.RootID, "body": item.Body, "spoiler": item.Spoiler, "hidden": item.Hidden, "created_at": item.CreatedAt, "replies": item.Replies, "reply_cursor": item.ReplyCursor}
}
func (handler *CommunityHandler) replies(request *gin.Context) {
	result, failure := handler.service.Replies(request.Request.Context(), request.Param("id"), request.Query("after"))
	if failure != nil {
		communityError(request, failure)
		return
	}
	items := []gin.H{}
	for _, item := range result.Comments {
		items = append(items, commentView(item))
	}
	request.JSON(200, gin.H{"comments": items, "next": result.Next})
}
func (handler *CommunityHandler) comment(request *gin.Context) {
	var input struct {
		TitleID   string `json:"title_id"`
		ReleaseID string `json:"release_id"`
		ParentID  string `json:"parent_id"`
		Body      string `json:"body"`
		Spoiler   bool   `json:"spoiler"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	id, failure := handler.service.Comment(request.Request.Context(), account.ID, domain.Comment{TitleID: input.TitleID, ReleaseID: input.ReleaseID, ParentID: input.ParentID, Body: input.Body, Spoiler: input.Spoiler})
	if failure != nil {
		communityError(request, failure)
		return
	}
	request.JSON(201, gin.H{"id": id})
}
func (handler *CommunityHandler) edit(request *gin.Context) {
	var input struct {
		Body    string `json:"body"`
		Spoiler bool   `json:"spoiler"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Edit(request.Request.Context(), account.ID, request.Param("id"), input.Body, input.Spoiler); failure != nil {
		communityError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *CommunityHandler) remove(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if failure := handler.service.Delete(request.Request.Context(), account.ID, request.Param("id")); failure != nil {
		communityError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *CommunityHandler) rating(request *gin.Context) {
	var input struct {
		Score int `json:"score"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Rating(request.Request.Context(), account.ID, request.Param("id"), input.Score); failure != nil {
		communityError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *CommunityHandler) myRating(request *gin.Context) {
	account, _ := CurrentAccount(request)
	score, failure := handler.service.MyRating(request.Request.Context(), account.ID, request.Param("id"))
	if failure != nil {
		communityError(request, failure)
		return
	}
	request.JSON(200, gin.H{"score": score})
}
func (handler *CommunityHandler) report(request *gin.Context) {
	var input struct {
		TitleID    string `json:"title_id"`
		ReleaseID  string `json:"release_id"`
		PageID     string `json:"page_id"`
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Category   string `json:"category"`
		Reason     string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	id, failure := handler.service.Report(request.Request.Context(), account.ID, application.Report{TitleID: input.TitleID, ReleaseID: input.ReleaseID, PageID: input.PageID, TargetType: input.TargetType, TargetID: input.TargetID, Category: input.Category, Reason: input.Reason})
	if failure != nil {
		communityError(request, failure)
		return
	}
	request.JSON(201, gin.H{"id": id})
}
func (handler *CommunityHandler) reports(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.Reports(request.Request.Context(), account.ID, request.Query("after"), request.Query("team"), request.Query("kind"))
	if failure != nil {
		communityError(request, failure)
		return
	}
	result := []gin.H{}
	for _, item := range items {
		result = append(result, gin.H{"id": item.ID, "title_id": item.TitleID, "release_id": item.ReleaseID, "page_id": item.PageID, "target_type": item.TargetType, "target_id": item.TargetID, "category": item.Category, "reason": item.Reason, "contact_email": item.ContactEmail, "status": item.Status, "created_at": item.CreatedAt,
			"title_name": item.TitleName, "title_slug": item.TitleSlug, "team_name": item.TeamName, "team_slug": item.TeamSlug, "chapter_label": item.ChapterLabel, "release_status": item.ReleaseStatus, "revision_id": item.RevisionID, "page_position": item.PagePosition, "comment_body": item.CommentBody, "comment_author": item.CommentAuthor, "comment_hidden": item.CommentHidden, "comment_deleted": item.CommentDeleted, "history": item.History})
	}
	next := ""
	if len(items) == 50 {
		next = items[len(items)-1].ID
	}
	request.Header("Cache-Control", "no-store")
	request.JSON(200, gin.H{"items": result, "has_more": next != "", "next": next})
}
func (handler *CommunityHandler) resolve(request *gin.Context) {
	var input struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Resolve(request.Request.Context(), account.ID, request.Param("id"), input.Status, input.Reason); failure != nil {
		communityError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *CommunityHandler) visibility(request *gin.Context) {
	var input struct {
		Hidden bool   `json:"hidden"`
		Reason string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Visibility(request.Request.Context(), account.ID, request.Param("id"), input.Hidden, input.Reason); failure != nil {
		communityError(request, failure)
		return
	}
	request.Status(204)
}
func communityError(request *gin.Context, failure error) {
	switch {
	case errors.Is(failure, domain.ErrInvalid):
		Error(request, 422, "validation", "Мәтін мен таңдалған нысанды тексеріңіз. Себеп кемінде 8 таңбадан тұруы керек.")
	case errors.Is(failure, domain.ErrForbidden):
		Error(request, 403, "forbidden", "Поштаңызды растаңыз және рұқсатты тексеріңіз")
	case errors.Is(failure, domain.ErrNotFound):
		Error(request, 404, "not_found", "Нысан қолжетімсіз")
	case errors.Is(failure, domain.ErrLimited):
		request.Header("Retry-After", "60")
		Error(request, 429, "rate_limit", "Тым жиі жіберілді. Кейінірек қайталаңыз.")
	default:
		log.Printf("community failure: %T", failure)
		Error(request, 503, "unavailable", "Қызмет уақытша қолжетімсіз")
	}
}
func (handler *CommunityHandler) rightsReport(request *gin.Context) {
	var input struct {
		TitleSlug string `json:"title_slug"`
		Email     string `json:"email"`
		Reason    string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	receipt, failure := handler.service.RightsReport(request.Request.Context(), application.RightsRequest{TitleSlug: input.TitleSlug, Email: input.Email, Reason: input.Reason})
	if failure != nil {
		communityError(request, failure)
		return
	}
	request.Header("Cache-Control", "no-store")
	request.JSON(201, gin.H{"id": receipt.ID, "token": receipt.Token})
}
func (handler *CommunityHandler) rightsStatus(request *gin.Context) {
	var input struct {
		ID    string `json:"id"`
		Token string `json:"token"`
	}
	if !decode(request, &input) {
		return
	}
	result, failure := handler.service.RightsStatus(request.Request.Context(), input.ID, input.Token)
	if failure != nil {
		communityError(request, failure)
		return
	}
	request.Header("Cache-Control", "no-store")
	request.JSON(200, gin.H{"id": result.ID, "status": result.Status, "resolution": result.Resolution})
}
func (handler *CommunityHandler) myReports(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.MyReports(request.Request.Context(), account.ID, request.Query("after"))
	if failure != nil {
		communityError(request, failure)
		return
	}
	result := []gin.H{}
	for _, item := range items {
		result = append(result, gin.H{"id": item.ID, "category": item.Category, "reason": item.Reason, "status": item.Status, "resolution": item.Resolution, "created_at": item.CreatedAt})
	}
	request.JSON(200, gin.H{"items": result, "has_more": len(items) == 50})
}
