package almanac

import (
	"errors"
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) GetList(c *gin.Context) {
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	items, total, err := ctrl.service.GetList(filter.Limit, filter.Offset)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch almanacs", err)
		return
	}
	utils.PaginatedOK(c, items, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	item, err := ctrl.service.GetById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Almanac not found")
		} else {
			utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch almanac", err)
		}
		return
	}
	utils.OK(c, item)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var req dto.AlmanacCreateDTO
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.Create(&req); err != nil {
		utils.FailInternal(c, "CREATE_ERROR", "Failed to create almanac", err)
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.AlmanacUpdateDTO
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.Update(id, &req); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Almanac not found")
		} else {
			utils.FailInternal(c, "UPDATE_ERROR", "Failed to update almanac", err)
		}
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Almanac not found")
		} else {
			utils.FailInternal(c, "DELETE_ERROR", "Failed to delete almanac", err)
		}
		return
	}
	utils.OK(c, nil)
}
