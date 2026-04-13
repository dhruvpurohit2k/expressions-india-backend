package team

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
	teams, err := ctrl.service.GetList()
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch teams", err)
		return
	}
	utils.OK(c, teams)
}

func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	team, err := ctrl.service.GetById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Team not found")
		} else {
			utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch team", err)
		}
		return
	}
	utils.OK(c, team)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var req dto.TeamCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid input")
		return
	}
	if err := ctrl.service.Create(&req); err != nil {
		utils.FailInternal(c, "CREATE_ERROR", "Failed to create team", err)
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.TeamUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid input")
		return
	}
	if err := ctrl.service.Update(id, &req); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Team not found")
		} else {
			utils.FailInternal(c, "UPDATE_ERROR", "Failed to update team", err)
		}
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Team not found")
		} else {
			utils.FailInternal(c, "DELETE_ERROR", "Failed to delete team", err)
		}
		return
	}
	utils.OK(c, nil)
}
