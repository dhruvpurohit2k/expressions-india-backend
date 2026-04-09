package course

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

func (ctrl *Controller) GetCoursesList(c *gin.Context) {
	courses, err := ctrl.service.GetCoursesList()
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch courses", err)
		return
	}
	utils.OK(c, courses)
}

func (ctrl *Controller) GetCourseById(c *gin.Context) {
	id := c.Param("id")
	course, err := ctrl.service.GetCourseById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Course not found")
		} else {
			utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch course", err)
		}
		return
	}
	utils.OK(c, course)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var req dto.CourseCreateRequestDTO
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.CreateCourse(&req); err != nil {
		utils.FailInternal(c, "CREATE_ERROR", "Failed to create course", err)
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.CourseCreateRequestDTO
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.UpdateCourse(id, &req); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Course not found")
		} else {
			utils.FailInternal(c, "UPDATE_ERROR", "Failed to update course", err)
		}
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.DeleteCourse(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Course not found")
		} else {
			utils.FailInternal(c, "DELETE_ERROR", "Failed to delete course", err)
		}
		return
	}
	utils.OK(c, nil)
}
