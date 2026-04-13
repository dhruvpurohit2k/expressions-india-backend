package journal

import (
	"errors"
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	JournalService *Service
}

func NewController(journalService *Service) *Controller {
	return &Controller{JournalService: journalService}
}

func (ctrl *Controller) GetList(c *gin.Context) {
	var filter utils.JournalFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	journals, total, err := ctrl.JournalService.GetJournalListFiltered(filter)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrieve journals: "+err.Error())
		return
	}
	utils.PaginatedOK(c, journals, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	journal, err := ctrl.JournalService.GetJournalById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Journal not found")
		} else {
			utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch journal")
		}
		return
	}
	utils.OK(c, journal)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.JournalService.DeleteJournal(id); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete journal")
		return
	}
	utils.OK(c, nil)
}

// GetAll returns every journal with full chapter data (used by public API).
func (ctrl *Controller) GetAll(c *gin.Context) {
	journals, err := ctrl.JournalService.GetAllJournals()
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch journals")
		return
	}
	if len(journals) == 0 {
		utils.OK(c, &[]dto.JournalListItemDTO{})
		return
	}
	utils.OK(c, journals)
}
