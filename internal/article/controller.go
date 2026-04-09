package article

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

func (ctrl *Controller) GetArticleList(c *gin.Context) {
	var filter utils.ArticleFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	articles, total, err := ctrl.service.GetArticleList(filter)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch articles", err)
		return
	}
	utils.PaginatedOK(c, articles, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetArticleListPaginated(c *gin.Context) {
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	articles, total, err := ctrl.service.GetArticleListPaginated(filter.Limit, filter.Offset)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch articles", err)
		return
	}
	utils.PaginatedOK(c, articles, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetArticlesByAudience(c *gin.Context) {
	audience := c.Param("audience")
	var filter utils.Filter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", err.Error())
		return
	}
	articles, total, err := ctrl.service.GetArticlesByAudience(audience, filter.Limit, filter.Offset)
	if err != nil {
		utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch articles", err)
		return
	}
	utils.PaginatedOK(c, articles, utils.Meta{
		Total:      total,
		PerPage:    filter.Limit,
		TotalPages: utils.SafeTotalPages(total, filter.Limit),
	})
}

func (ctrl *Controller) GetArticleById(c *gin.Context) {
	id := c.Param("id")
	article, err := ctrl.service.GetArticleById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, http.StatusNotFound, "NOT_FOUND", "Article not found")
		} else {
			utils.FailInternal(c, "FETCH_ERROR", "Failed to fetch article", err)
		}
		return
	}
	utils.OK(c, article)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var req dto.ArticleCreateRequestDTO
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.CreateArticle(&req); err != nil {
		utils.FailInternal(c, "CREATE_ERROR", "Failed to create article", err)
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.service.DeleteArticle(id); err != nil {
		utils.FailInternal(c, "DELETE_ERROR", "Failed to delete article", err)
		return
	}
	utils.OK(c, nil)
}

func (ctrl *Controller) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.ArticleUpdateRequestDTO
	if err := c.ShouldBind(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}
	if err := ctrl.service.UpdateArticle(id, &req); err != nil {
		utils.FailInternal(c, "UPDATE_ERROR", "Failed to update article", err)
		return
	}
	utils.OK(c, nil)
}
