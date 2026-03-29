package event

import (
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service
}

func NewController(s Service) *Controller {
	return &Controller{
		service: s,
	}
}

func (c *Controller) GetAll(ctx *gin.Context) {

	events, err := c.service.GetAllEvents()

	if err != nil {
		utils.Fail(ctx, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrive events")
	}

	utils.OK(ctx, events)

}
