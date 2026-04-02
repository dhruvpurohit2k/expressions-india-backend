package event

import (
	"fmt"
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

	// ctrl.service.CreateEvent(event, images, audiences)
	// c.JSON(http.StatusOK, gin.H{"message": "Event created successfully"})
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

// func (ctrl *Controller) getEventFromForm(c *gin.Context) (*models.Event, []*multipart.FileHeader, []string, error) {
// 	title := c.PostForm("title")
// 	description := c.PostForm("description")
// 	location := c.PostForm("location")
// 	isPaid := c.PostForm("isPaid") == "true"

// 	price, err := strconv.Atoi(c.PostForm("price"))
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	startDate, err := time.Parse(time.RFC3339, c.PostForm("startDate"))
// 	if err != nil {
// 		utils.Fail(c, http.StatusBadRequest, "INVALID_DATE", "Date must be YYYY-MM-DD")
// 		return nil, nil, nil, err
// 	}
// 	endDate, err := time.Parse(time.RFC3339, c.PostForm("endDate"))
// 	if err != nil {
// 		utils.Fail(c, http.StatusBadRequest, "INVALID_DATE", "Date must be YYYY-MM-DD")
// 		return nil, nil, nil, err
// 	}
// 	form, _ := c.MultipartForm()
// 	images := form.File["medias"]
// 	perks := []byte(c.PostForm("perks"))
// 	var audiences []string
// 	audiencesRaw := c.PostForm("audiences")
// 	json.Unmarshal([]byte(audiencesRaw), &audiences)
// 	event := &models.Event{
// 		Title:       title,
// 		Description: description,
// 		StartDate:   startDate,
// 		EndDate:     &endDate,
// 		Location:    &location,
// 		IsPaid:      isPaid,
// 		Price:       &price,
// 		Perks:       datatypes.JSON(perks),
// 		Audiences:   []models.Audience{},
// 		Medias:      []models.Media{},
// 	}

// 	return event, images, audiences, nil
// }
