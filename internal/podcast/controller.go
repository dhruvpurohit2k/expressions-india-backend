package podcast

import (
	"math"
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

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.DeletePodcast(id); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete podcast")
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) GetPodcastList(c *gin.Context) {
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	podcasts, total, err := ctrl.service.GetPodcastList(filter.Limit, filter.Offset)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrieve podcasts: "+err.Error())
		return
	}
	utils.PaginatedOK(c, podcasts, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	})
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
