package audience

import (
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"errors"
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

func (ctrl *Controller) UpdateDescription(c *gin.Context) {
	id := c.Param("id")
	var req dto.AudienceUpdateDescriptionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.UpdateAudienceDescription(id, req.Description); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Audience not found")
		} else {
			utils.Fail(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update audience: "+err.Error())
		}
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) GetAudience(c *gin.Context) {
	audience, err := ctrl.service.GetAudience(c)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch audience")
		return
	}
	utils.OK(c, audience)
}
