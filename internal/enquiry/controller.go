package enquiry

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
	var filter utils.EnquiryFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	enquiries, total, err := ctrl.service.GetEnquiryListFiltered(filter)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Error fetching enquiries", err)
		return
	}
	utils.PaginatedOK(c, enquiries, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	enquiry, err := ctrl.service.GetEnquiryById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Enquiry not found")
		} else {
			utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Error fetching enquiry")
		}
		return
	}
	utils.OK(c, enquiry)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.DeleteEnquiry(id); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "DELETE_ERROR", "Error deleting enquiry")
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) CreateEnquiry(c *gin.Context) {
	var enquiry dto.EnquiryCreateDTO
	if err := c.ShouldBind(&enquiry); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "Error binding enquiry data")
		return
	}
	if err := ctrl.service.CreateEnquiry(&enquiry); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "CREATE_ERROR", "Error creating enquiry")
		return
	}
	utils.OK(c, gin.H{"message": "Enquiry created successfully"})
}
