package event

import (
	"fmt"
	"math"
	"net/http"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service
}

func NewController(s Service) *Controller {
	return &Controller{
		service: s,
	}
}

func (ctrl *Controller) GetAll(c *gin.Context) {
	events, err := ctrl.service.GetAllEvents()
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrive events: "+err.Error())
		return
	}
	utils.OK(c, events)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var newEvent dto.EventCreateRequestDTO
	if err := c.ShouldBind(&newEvent); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.CreateEvent(&newEvent); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "CREATE_ERROR", "Could not create event: "+err.Error())
		return
	}
	utils.OK(c, &newEvent)

}

func (ctrl *Controller) GetEventList(c *gin.Context) {
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	events, err := ctrl.service.GetEventList(filter)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrive events: "+err.Error())
		return
	}
	fmt.Println(events)
	if len(events) == 0 {
		utils.OK(c, &[]dto.EventListItemDTO{})
		return
	}
	utils.OK(c, events)
}

func (ctrl *Controller) Update(c *gin.Context) {
	id := c.Param("id")
	var updateEvent dto.EventUpdateRequestDTO
	if err := c.ShouldBind(&updateEvent); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.UpdateEvent(id, &updateEvent); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}
	utils.OK(c, updateEvent)

}
func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.DeleteEvent(id); err != nil {
		utils.Fail(c, http.StatusInternalServerError, "DELETE_ERROR", "Could not delete event: "+err.Error())
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) GetEventById(c *gin.Context) {
	id := c.Param("id")
	event, err := ctrl.service.GetEventById(id)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrieve event: "+err.Error())
		return
	}
	utils.OK(c, event)
}

func (ctrl *Controller) GetUpcomingEventsByAudience(c *gin.Context) {
	audience := c.Param("audience")
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	events, total, err := ctrl.service.GetUpcomingEventsByAudience(audience, filter.Limit, filter.Offset)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrieve events: "+err.Error())
		return
	}
	utils.PaginatedOK(c, events, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	})
}

func (ctrl *Controller) GetUpcomingEvents(c *gin.Context) {
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	events, total, err := ctrl.service.GetUpcomingEvents(filter.Limit, filter.Offset)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrieve upcoming events: "+err.Error())
		return
	}
	utils.PaginatedOK(c, events, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	})
}

func (ctrl *Controller) GetPastEvents(c *gin.Context) {
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	events, total, err := ctrl.service.GetPastEvents(filter.Limit, filter.Offset)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "FETCH_ERROR", "Could not retrieve past events: "+err.Error())
		return
	}
	utils.PaginatedOK(c, events, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	})
}
