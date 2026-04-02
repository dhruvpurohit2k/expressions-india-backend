package enquiry

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
	enquiries, err := ctrl.service.GetEnquiryList()
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Error fetching enquiries")
		return
	}
	if len(enquiries) == 0 {
		utils.OK(c, &[]dto.EnquiryListItemDTO{})
		return
	}
	utils.OK(c, enquiries)
}

func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	enquiry, err := ctrl.service.GetEnquiryById(id)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Error fetching enquiry")
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
