package promotion

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

func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	promotion, err := ctrl.service.GetById(id)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Error fetching promotion")
		return
	}
	utils.OK(c, promotion)
}

func (ctrl *Controller) Get(c *gin.Context) {
	promotions, err := ctrl.service.Get()
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Error fetching promotions")
		return
	}
	utils.OK(c, promotions)
	return
}
