package list

import (
	"PocketArtisan/internal/http/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func RegisterRoutes(router *gin.RouterGroup, mongoDB *mongo.Database) {
	uc := NewService(mongoDB)

	router.GET("/comment/:productID", func(c *gin.Context) {
		productID, err := strconv.ParseUint(c.Param("productID"), 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid productID")
			return
		}

		skip, err := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid skip")
			return
		}
		limit, err := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid limit")
			return
		}

		resp, err := uc.Execute(c.Request.Context(), Request{ProductID: productID, Skip: skip, Limit: limit})
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Data(c, http.StatusOK, resp)
	})
}
