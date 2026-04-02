package journal

import (
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	JournalService *Service
}

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

func (ctrl *Controller) GetList(c *gin.Context) {

	journals, err := ctrl.JournalService.Get()
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
func (ctrl *Controller) GetById(c *gin.Context) {
	id := c.Param("id")
	journal, err := ctrl.JournalService.GetJournalById(id)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch journal")
		return
	}
	utils.OK(c, journal)
}

func NewController(journalService *Service) *Controller {
	return &Controller{JournalService: journalService}
}
