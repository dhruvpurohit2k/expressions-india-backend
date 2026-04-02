package podcast

import (
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) Get(c *gin.Context) {
	podcasts, err := ctrl.service.GetPodcasts()
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch podcasts")
		return
	}
	if len(podcasts) == 0 {
		utils.OK(c, &[]dto.PodcastDTO{})
		return
	}
	utils.OK(c, podcasts)
}

func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	podcast, err := ctrl.service.GetPodcastById(id)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch podcast")
		return
	}
	if podcast == nil {
		utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Podcast not found")
		return
	}
	utils.OK(c, podcast)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var dto dto.PodcastCreateDTO
	if err := c.ShouldBind(&dto); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid input")
		return
	}
	if err := ctrl.service.CreatePodcast(&dto); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create podcast")
		return
	}
	utils.OK(c, nil)
}
