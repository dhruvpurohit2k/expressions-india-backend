package audience

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

func (ctrl *Controller) GetAudienceByName(c *gin.Context) {
	name := c.Param("name")
	audience, err := ctrl.service.GetAudienceByName(name)
	if err != nil {
		utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Audience not found")
		return
	}
	utils.OK(c, audience)
}

func (ctrl *Controller) GetAudience(c *gin.Context) {
	audience, err := ctrl.service.GetAudience(c)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch audience")
		return
	}
	utils.OK(c, audience)
}
