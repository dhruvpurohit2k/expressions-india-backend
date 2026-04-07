package article

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/storage"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
	s3 *storage.S3
}

func NewService(db *gorm.DB, s3 *storage.S3) *Service {
	return &Service{db: db, s3: s3}
}

func (s *Service) GetArticleList(filter utils.ArticleFilter) ([]dto.ArticleListItemDTO, int64, error) {
	var articles []models.Article
	var total int64

	base := s.db.Model(&models.Article{})
	if filter.Search != "" {
		base = base.Where("title LIKE ?", "%"+filter.Search+"%")
	}
	if filter.Category != "" {
		base = base.Where("category LIKE ?", "%"+filter.Category+"%")
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "created_at DESC"
	if filter.SortOrder == "asc" {
		order = "created_at ASC"
	}

	if err := base.Order(order).Limit(filter.Limit).Offset(filter.Offset).Preload("Medias").Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.ArticleListItemDTO, 0, len(articles))
	for _, article := range articles {
		item := dto.ArticleListItemDTO{
			ID:          article.ID,
			Title:       article.Title,
			Category:    article.Category,
			PublishedAt: article.CreatedAt,
		}
		if len(article.Medias) > 0 {
			item.ThumbnailURL = &article.Medias[0].URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *Service) GetArticleListPaginated(limit int, offset int) ([]dto.ArticleListItemDTO, int64, error) {
	var articles []models.Article
	var total int64

	base := s.db.Model(&models.Article{})
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := s.db.Model(&models.Article{}).Order("created_at DESC").Limit(limit).Offset(offset).Preload("Medias").Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.ArticleListItemDTO, 0, len(articles))
	for _, article := range articles {
		item := dto.ArticleListItemDTO{
			ID:          article.ID,
			Title:       article.Title,
			Category:    article.Category,
			PublishedAt: article.CreatedAt,
		}
		if len(article.Medias) > 0 {
			item.ThumbnailURL = &article.Medias[0].URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *Service) GetArticlesByAudience(audience string, limit int, offset int) ([]dto.ArticleListItemDTO, int64, error) {
	var articles []models.Article
	var total int64

	base := s.db.Model(&models.Article{}).
		Where(
			"articles.id IN (SELECT aa.article_id FROM article_audience aa JOIN audiences a ON a.id = aa.audience_id WHERE a.name = ? OR a.name = 'all') OR articles.id NOT IN (SELECT DISTINCT aa.article_id FROM article_audience aa)",
			audience,
		)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := base.Order("articles.created_at DESC").Limit(limit).Offset(offset).Preload("Medias").Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.ArticleListItemDTO, 0, len(articles))
	for _, a := range articles {
		item := dto.ArticleListItemDTO{
			ID:          a.ID,
			Title:       a.Title,
			Category:    a.Category,
			PublishedAt: a.CreatedAt,
		}
		if len(a.Medias) > 0 {
			item.ThumbnailURL = &a.Medias[0].URL
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *Service) GetArticleById(id string) (*dto.ArticleDTO, error) {
	var article models.Article
	if err := s.db.Where("id = ?", id).
		Preload("Audience").
		Preload("Medias").
		First(&article).Error; err != nil {
		return nil, err
	}
	return &dto.ArticleDTO{
		ID:        article.ID,
		Title:     article.Title,
		Content:   article.Content,
		Category:  article.Category,
		Audience:  article.Audience,
		Medias:    article.Medias,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}, nil
}

func (s *Service) CreateArticle(data *dto.ArticleCreateRequestDTO) (retErr error) {
	article := models.Article{
		Title:    data.Title,
		Content:  data.Content,
		Category: data.Category,
	}

	if len(data.Audiences) > 0 {
		audiences, err := s.resolveAudiences(data.Audiences)
		if err != nil {
			return err
		}
		article.Audience = audiences
	}

	if len(data.Medias) > 0 {
		medias, err := s.uploadMediaFiles(data.Medias)
		if err != nil {
			return err
		}
		article.Medias = medias
	}

	if err := s.db.Create(&article).Error; err != nil {
		// Clean up uploaded S3 files since the DB record failed
		for _, m := range article.Medias {
			if delErr := s.s3.Delete(m.ID); delErr != nil {
				log.Printf("S3 cleanup failed for key %s: %v", m.ID, delErr)
			}
		}
		return err
	}
	return nil
}

func (s *Service) UpdateArticle(id string, data *dto.ArticleUpdateRequestDTO) error {
	var article models.Article
	if err := s.db.Where("id = ?", id).First(&article).Error; err != nil {
		return err
	}
	article.Title = data.Title
	article.Content = data.Content
	article.Category = data.Category

	if len(data.Audiences) > 0 {
		audiences, err := s.resolveAudiences(data.Audiences)
		if err != nil {
			return err
		}
		if err := s.db.Model(&article).Association("Audience").Replace(audiences); err != nil {
			return err
		}
	} else {
		if err := s.db.Model(&article).Association("Audience").Clear(); err != nil {
			return err
		}
	}

	for _, mediaId := range data.DeletedMediaIds {
		if err := s.db.Model(&article).Association("Medias").Unscoped().Delete(&models.Media{ID: mediaId}); err != nil {
			return err
		}
		if err := s.db.Delete(&models.Media{}, "id = ?", mediaId).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(mediaId); err != nil {
			log.Printf("failed to delete media %s from S3: %v", mediaId, err)
		}
	}

	if len(data.Medias) > 0 {
		medias, err := s.uploadMediaFiles(data.Medias)
		if err != nil {
			return err
		}
		if err := s.db.Model(&article).Association("Medias").Append(medias); err != nil {
			// Clean up uploaded S3 files since the DB association failed
			for _, m := range medias {
				if delErr := s.s3.Delete(m.ID); delErr != nil {
					log.Printf("S3 cleanup failed for key %s: %v", m.ID, delErr)
				}
			}
			return err
		}
	}

	return s.db.Save(&article).Error
}

func (s *Service) DeleteArticle(id string) error {
	var article models.Article
	if err := s.db.Where("id = ?", id).Preload("Medias").First(&article).Error; err != nil {
		return err
	}
	if err := s.db.Model(&article).Association("Audience").Clear(); err != nil {
		return err
	}
	for _, media := range article.Medias {
		if err := s.db.Model(&article).Association("Medias").Delete(&media); err != nil {
			return err
		}
		if err := s.db.Delete(&models.Media{}, "id = ?", media.ID).Error; err != nil {
			return err
		}
		if err := s.s3.Delete(media.ID); err != nil {
			log.Printf("failed to delete media %s from S3: %v", media.ID, err)
		}
	}
	return s.db.Delete(&article).Error
}

// resolveAudiences fetches audience records matching the given names in a single query.
func (s *Service) resolveAudiences(names []string) ([]models.Audience, error) {
	var audiences []models.Audience
	if err := s.db.Where("name IN ?", names).Find(&audiences).Error; err != nil {
		return nil, err
	}
	return audiences, nil
}

// uploadMediaFiles uploads all files to S3 concurrently.
// If any upload fails, successfully uploaded files are cleaned up before returning the error.
func (s *Service) uploadMediaFiles(files []*multipart.FileHeader) ([]models.Media, error) {
	medias := make([]models.Media, len(files))
	g, _ := errgroup.WithContext(context.Background())
	for i, file := range files {
		i, file := i, file
		g.Go(func() error {
			f, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", file.Filename, err)
			}
			defer f.Close()
			location, key, contentType, err := s.s3.UploadNetwork(f)
			if err != nil {
				return err
			}
			medias[i] = models.Media{ID: key, URL: location, FileType: contentType}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		for _, m := range medias {
			if m.ID != "" {
				if delErr := s.s3.Delete(m.ID); delErr != nil {
					log.Printf("S3 cleanup failed for key %s: %v", m.ID, delErr)
				}
			}
		}
		return nil, err
	}
	return medias, nil
}
