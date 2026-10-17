package http

import "github.com/gin-gonic/gin"

func (handler *TeamHandler) reject(request *gin.Context) {
	var input struct {
		Reason string `json:"reason"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Reject(request.Request.Context(), account.ID, request.Param("id"), input.Reason); failure != nil {
		teamError(request, failure)
		return
	}
	request.Status(204)
}
func (handler *TeamHandler) update(request *gin.Context) {
	var input struct {
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !decode(request, &input) {
		return
	}
	account, _ := CurrentAccount(request)
	if failure := handler.service.Update(request.Request.Context(), account.ID, request.Param("id"), input.Slug, input.Name, input.Description); failure != nil {
		teamError(request, failure)
		return
	}
	request.Status(204)
}
