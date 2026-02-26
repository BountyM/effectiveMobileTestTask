package handler

import "github.com/gin-gonic/gin"

type errorResponse struct {
	Error string `json:"error"`
}

func newErrorResponse(c *gin.Context, statusCode int, err string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{err})
}
