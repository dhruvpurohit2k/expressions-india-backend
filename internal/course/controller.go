package course

import (
	"errors"
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/auth"
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

func (ctrl *Controller) GetCoursesListAdmin(c *gin.Context) {
	var filter utils.CourseFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	courses, total, err := ctrl.service.GetCoursesListFiltered(filter)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch courses", err)
		return
	}
	utils.PaginatedOK(c, courses, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetCoursesList(c *gin.Context) {
	var filter utils.CourseFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	courses, total, err := ctrl.service.GetCoursesListFiltered(filter)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch courses", err)
		return
	}
	utils.PaginatedOK(c, courses, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
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

func (ctrl *Controller) GetChapterById(c *gin.Context) {
	courseId := c.Param("id")
	chapterId := c.Param("chapterId")

	chapter, err := ctrl.service.GetChapterById(courseId, chapterId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Chapter not found")
		} else {
			utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch chapter", err)
		}
		return
	}

	if !chapter.IsFree {
		claims := auth.GetClaims(c)
		if claims == nil {
			utils.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "login required to access this chapter")
			return
		}
		if !claims.IsAdmin {
			enrolled, err := ctrl.service.IsEnrolled(courseId, claims.UserID)
			if err != nil {
				utils.FailInternal(c, "ENROLLMENT_CHECK_ERROR", "Failed to check enrollment", err)
				return
			}
			if !enrolled {
				utils.Fail(c, http.StatusForbidden, "FORBIDDEN", "purchase this course to access this chapter")
				return
			}
		}
	}

	utils.OK(c, chapter)
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

func (ctrl *Controller) GetCoursesByAudience(c *gin.Context) {
	audience := c.Param("audience")
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	courses, total, err := ctrl.service.GetCoursesByAudience(audience, filter.Limit, filter.Offset)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Could not retrieve courses", err)
		return
	}
	utils.PaginatedOK(c, courses, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
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

func (ctrl *Controller) GetEnrolledUsers(c *gin.Context) {
	courseId := c.Param("id")
	var filter utils.EnrollmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	users, total, err := ctrl.service.GetEnrolledUsers(courseId, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Course not found")
		} else {
			utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch enrolled users", err)
		}
		return
	}
	utils.PaginatedOK(c, users, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetNotEnrolledUsers(c *gin.Context) {
	courseId := c.Param("id")
	var filter utils.EnrollmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	users, total, err := ctrl.service.GetNotEnrolledUsers(courseId, filter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Course not found")
		} else {
			utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch non-enrolled users", err)
		}
		return
	}
	utils.PaginatedOK(c, users, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) EnrollUser(c *gin.Context) {
	courseId := c.Param("id")
	var req dto.EnrollUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "userId is required")
		return
	}
	if err := ctrl.service.EnrollUser(courseId, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Course or user not found")
		} else {
			utils.FailInternal(c, "ENROLL_ERROR", "Failed to enroll user", err)
		}
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) RevokeAccess(c *gin.Context) {
	courseId := c.Param("id")
	userId := c.Param("userId")
	if err := ctrl.service.RevokeAccess(courseId, userId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Course not found")
		} else {
			utils.FailInternal(c, "REVOKE_ERROR", "Failed to revoke access", err)
		}
		return
	}
	utils.OK(c, nil)
}
