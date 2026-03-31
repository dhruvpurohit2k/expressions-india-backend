package journal

import (
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	JournalService *Service
}

func (ctrl *Controller) Get(c *gin.Context) {

	journals, err := ctrl.JournalService.GetJournals()
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch journals")
		return
	}
	utils.OK(c, journals)
}

func NewController(journalService *Service) *Controller {
	return &Controller{JournalService: journalService}
}
