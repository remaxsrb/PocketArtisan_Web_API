package response

import "github.com/gin-gonic/gin"

type ErrorPayload struct {
	Error string `json:"error"`
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorPayload{Error: message})
}

func Data(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{"data": data})
}
