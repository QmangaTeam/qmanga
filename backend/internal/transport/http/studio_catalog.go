package http

import (
	"github.com/gin-gonic/gin"
	"qmanga/internal/catalog/domain"
)

func projectRequestView(item domain.ProjectRequest) gin.H {
	return gin.H{"id": item.ID, "team_id": item.TeamID, "team_name": item.TeamName, "title_id": item.TitleID, "title_name": item.TitleName, "name": item.Name, "original_name": item.OriginalName, "slug": item.Slug, "kind": item.Kind, "description": item.Description, "source_url": item.SourceURL, "note": item.Note, "status": item.Status, "resolution": item.Resolution, "project_id": item.ProjectID, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
}
func (handler *CatalogHandler) projectRequests(request *gin.Context) {
	account, _ := CurrentAccount(request)
	items, failure := handler.service.ProjectRequests(request.Request.Context(), account.ID, request.FullPath() == "/api/v1/admin/project-requests", request.Query("after"))
	if failure != nil {
		catalogError(request, failure)
		return
	}
	views := []gin.H{}
	for _, item := range items {
		views = append(views, projectRequestView(item))
	}
	request.JSON(200, gin.H{"items": views, "has_more": len(items) == 50})
}
func (handler *CatalogHandler) requestProject(request *gin.Context) {
	var input struct {
		TeamID       string `json:"team_id"`
		TitleID      string `json:"title_id"`
		Name         string `json:"name"`
		OriginalName string `json:"original_name"`
		Slug         string `json:"slug"`
		Kind         string `json:"kind"`
		Description  string `json:"description"`
		SourceURL    string `json:"source_url"`
		Note         string `json:"note"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	item, failure := handler.service.RequestProject(request.Request.Context(), account.ID, domain.ProjectRequestInput{TeamID: input.TeamID, TitleID: input.TitleID, Name: input.Name, OriginalName: input.OriginalName, Slug: input.Slug, Kind: input.Kind, Description: input.Description, SourceURL: input.SourceURL, Note: input.Note})
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(201, projectRequestView(item))
}
func (handler *CatalogHandler) reviewProjectRequest(request *gin.Context) {
	var input struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
		TitleID  string `json:"title_id"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	item, failure := handler.service.ReviewProjectRequest(request.Request.Context(), account.ID, request.Param("id"), input.Decision, input.Reason, input.TitleID)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(200, projectRequestView(item))
}
func (handler *CatalogHandler) cancelProjectRequest(request *gin.Context) {
	account, _ := CurrentAccount(request)
	if failure := handler.service.CancelProjectRequest(request.Request.Context(), account.ID, request.Param("id")); failure != nil {
		catalogError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *CatalogHandler) studioProject(request *gin.Context) {
	account, _ := CurrentAccount(request)
	item, failure := handler.service.StudioProject(request.Request.Context(), account.ID, request.Param("id"))
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(200, gin.H{"id": item.Project.ID, "title_id": item.Project.TitleID, "title_name": item.Project.TitleName, "team_id": item.Project.TeamID, "team_name": item.Project.TeamName, "role": item.Role, "approved": item.Approved, "status": item.Project.Status})
}
func (handler *CatalogHandler) studioCreateChapter(request *gin.Context) {
	var input struct {
		Label  string `json:"label"`
		Volume string `json:"volume"`
		Name   string `json:"name"`
		Order  int    `json:"order"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	id, failure := handler.service.CreateStudioChapter(request.Request.Context(), account.ID, request.Param("id"), input.Label, input.Volume, input.Name, input.Order)
	if failure != nil {
		catalogError(request, failure)
		return
	}
	request.JSON(201, gin.H{"id": id})
}
