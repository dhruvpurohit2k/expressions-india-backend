package upload

import (
	"net/http"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const presignTTL = 15 * time.Minute

type Controller struct {
	s3 *storage.S3
}

func NewController(s3 *storage.S3) *Controller {
	return &Controller{s3: s3}
}

// Presign accepts a list of {contentType, fileName} and returns a presigned
// PUT URL for each. The client uploads directly to S3, then passes the
// returned IDs in the final resource-creation form submit.
func (ctrl *Controller) Presign(c *gin.Context) {
	var items []dto.PresignRequestItem
	if err := c.ShouldBindJSON(&items); err != nil {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if len(items) == 0 {
		utils.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "at least one file is required")
		return
	}

	result := make([]dto.PresignResponseItem, 0, len(items))
	for _, item := range items {
		id := uuid.Must(uuid.NewV7()).String()
		presignedURL, err := ctrl.s3.PresignUpload(id, item.ContentType, presignTTL)
		if err != nil {
			utils.FailInternal(c, "PRESIGN_ERROR", "failed to generate presigned URL", err)
			return
		}
		result = append(result, dto.PresignResponseItem{
			ID:           id,
			PresignedURL: presignedURL,
			URL:          ctrl.s3.PublicURL(id),
			ContentType:  item.ContentType,
			FileName:     item.FileName,
		})
	}

	utils.OK(c, result)
}
