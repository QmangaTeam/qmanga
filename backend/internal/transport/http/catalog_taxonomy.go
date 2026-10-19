package http

import (
	"github.com/gin-gonic/gin"
	"qmanga/internal/catalog/domain"
)

type collectionDTO struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Position    int      `json:"position"`
	Visible     bool     `json:"visible"`
	TitleIDs    []string `json:"title_ids"`
}

func collectionViews(items []domain.Collection) []collectionDTO {
	result := []collectionDTO{}
	for _, item := range items {
		ids := item.TitleIDs
		if ids == nil {
			ids = []string{}
		}
		result = append(result, collectionDTO{item.ID, item.Name, item.Slug, item.Description, item.Position, item.Visible, ids})
	}
	return result
}
func (handler *CatalogHandler) taxonomy(request *gin.Context) {
	actor := ""
	if request.FullPath() == "/api/v1/admin/taxonomy" {
		account, _ := CurrentAccount(request)
		actor = account.ID
	}
	items, err := handler.service.Taxonomy(request.Request.Context(), actor)
	if err != nil {
		catalogError(request, err)
		return
	}
	request.JSON(200, gin.H{"genres": genreViews(items.Genres), "categories": genreViews(items.Categories), "collections": collectionViews(items.Collections)})
}
func (handler *CatalogHandler) saveTaxon(kind string) gin.HandlerFunc {
	return func(request *gin.Context) {
		var input struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if !decode(request, &input) {
			return
		}
		account, _ := CurrentAccount(request)
		item, err := handler.service.SaveTaxon(request.Request.Context(), account.ID, kind, request.Param("id"), input.Name, input.Slug)
		if err != nil {
			catalogError(request, err)
			return
		}
		request.JSON(200, genreDTO{ID: item.ID, Name: item.Name, Slug: item.Slug})
	}
}
func (handler *CatalogHandler) deleteTaxon(kind string) gin.HandlerFunc {
	return func(request *gin.Context) {
		account, _ := CurrentAccount(request)
		if err := handler.service.DeleteTaxon(request.Request.Context(), account.ID, kind, request.Param("id")); err != nil {
			catalogError(request, err)
			return
		}
		request.Status(204)
	}
}
func (handler *CatalogHandler) saveCollection(request *gin.Context) {
	var input collectionDTO
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	item, err := handler.service.SaveCollection(request.Request.Context(), account.ID, domain.Collection{ID: request.Param("id"), Name: input.Name, Slug: input.Slug, Description: input.Description, Position: input.Position, Visible: input.Visible, TitleIDs: input.TitleIDs})
	if err != nil {
		catalogError(request, err)
		return
	}
	request.JSON(200, collectionViews([]domain.Collection{item})[0])
}
func (handler *CatalogHandler) setCategories(request *gin.Context) {
	var input struct {
		IDs []string `json:"category_ids"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if err := handler.service.SetCategories(request.Request.Context(), account.ID, request.Param("id"), input.IDs); err != nil {
		catalogError(request, err)
		return
	}
	request.Status(204)
}
