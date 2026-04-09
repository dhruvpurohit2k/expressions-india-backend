package latestactivity

import (
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) GetLatestActivity(c *gin.Context) {
	activities, err := ctrl.service.GetLatestActivity()
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", err.Error())
		return
	}
	utils.OK(c, activities)
}
