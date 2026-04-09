package course

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/dto"
	"github.com/dhruvpurohit2k/expressions-india-backend/internal/models"
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

func (s *Service) GetCoursesList() ([]dto.CourseListItemDTO, error) {
	var courses []models.Course
	if err := s.db.Preload("Audiences").Find(&courses).Error; err != nil {
		return nil, err
	}
	courseList := make([]dto.CourseListItemDTO, 0, len(courses))
	for _, course := range courses {
		item := dto.CourseListItemDTO{
			ID:        course.ID,
			Title:     course.Title,
			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,
			Audiences: make([]string, 0, len(course.Audiences)),
		}
		for _, audience := range course.Audiences {
			item.Audiences = append(item.Audiences, audience.Name)
		}
		courseList = append(courseList, item)
	}
	return courseList, nil
}

func (s *Service) GetCourseById(id string) (*dto.CourseDTO, error) {
	var course models.Course
	if err := s.db.Where("id = ?", id).
		Preload("Audiences").
		Preload("Thumbnail").
		Preload("IntroductionVideo").
		Preload("DownloadableContent").
		Preload("Chapters").
		Preload("Chapters.DownloadableContent").
		Preload("Chapters.VideoLink").
		First(&course).Error; err != nil {
		return nil, err
	}

	result := &dto.CourseDTO{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		CreatedAt:   course.CreatedAt,
		UpdatedAt:   course.UpdatedAt,
	}

	if course.ThumbnailID != "" {
		result.Thumbnail = &dto.CourseMediaDTO{
			ID:       course.Thumbnail.ID,
			Name:     course.Thumbnail.Name,
			URL:      course.Thumbnail.URL,
			FileType: course.Thumbnail.FileType,
		}
	}

	if course.IntroductionVideoID != "" {
		result.IntroductionVideoURL = course.IntroductionVideo.URL
	}

	for _, a := range course.Audiences {
		result.Audiences = append(result.Audiences, a.Name)
	}

	for _, m := range course.DownloadableContent {
		result.DownloadableContent = append(result.DownloadableContent, dto.CourseMediaDTO{
			ID: m.ID, Name: m.Name, URL: m.URL, FileType: m.FileType,
		})
	}

	for _, ch := range course.Chapters {
		chDTO := dto.CourseChapterDTO{
			ID:          ch.ID,
			Title:       ch.Title,
			Description: ch.Description,
			IsFree:      ch.IsFree,
		}
		if ch.VideoLinkID != "" {
			chDTO.VideoLinkURL = ch.VideoLink.URL
		}
		for _, m := range ch.DownloadableContent {
			chDTO.DownloadableContent = append(chDTO.DownloadableContent, dto.CourseMediaDTO{
				ID: m.ID, Name: m.Name, URL: m.URL, FileType: m.FileType,
			})
		}
		result.Chapters = append(result.Chapters, chDTO)
	}

	return result, nil
}

func (s *Service) CreateCourse(data *dto.CourseCreateRequestDTO) error {
	chapters, err := data.ParsedChapters()
	if err != nil {
		return fmt.Errorf("invalid chaptersJson: %w", err)
	}

	course := models.Course{
		Title:       data.Title,
		Description: data.Description,
	}

	if len(data.Audiences) > 0 {
		audiences, err := s.resolveAudiences(data.Audiences)
		if err != nil {
			return err
		}
		course.Audiences = audiences
	}

	if data.Thumbnail != nil {
		media, err := s.uploadSingleFile(data.Thumbnail, "")
		if err != nil {
			return err
		}
		s.db.Create(&media)
		course.ThumbnailID = media.ID
	}

	if data.IntroductionVideoUrl != "" {
		link := models.Link{URL: data.IntroductionVideoUrl}
		if err := s.db.Create(&link).Error; err != nil {
			return err
		}
		course.IntroductionVideoID = link.ID
	}

	if err := s.db.Create(&course).Error; err != nil {
		return err
	}

	if len(data.DocFiles) > 0 {
		medias, err := s.uploadMediaFilesWithNames(data.DocFiles, data.DocNames)
		if err != nil {
			return err
		}
		if err := s.db.Model(&course).Association("DownloadableContent").Append(medias); err != nil {
			return err
		}
	}

	// Create chapters
	if err := s.createChapters(course.ID, chapters, data.ChapterDocFiles); err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateCourse(id string, data *dto.CourseCreateRequestDTO) error {
	var course models.Course
	if err := s.db.Where("id = ?", id).
		Preload("Thumbnail").
		Preload("IntroductionVideo").
		Preload("DownloadableContent").
		Preload("Chapters").
		Preload("Chapters.DownloadableContent").
		Preload("Chapters.VideoLink").
		First(&course).Error; err != nil {
		return err
	}

	chapters, err := data.ParsedChapters()
	if err != nil {
		return fmt.Errorf("invalid chaptersJson: %w", err)
	}

	course.Title = data.Title
	course.Description = data.Description

	// Handle thumbnail
	if data.DeletedThumbnailId != "" {
		if err := s.deleteMedia(data.DeletedThumbnailId); err != nil {
			log.Printf("failed to delete thumbnail %s: %v", data.DeletedThumbnailId, err)
		}
		course.ThumbnailID = ""
	}
	if data.Thumbnail != nil {
		media, err := s.uploadSingleFile(data.Thumbnail, "")
		if err != nil {
			return err
		}
		if err := s.db.Create(&media).Error; err != nil {
			return err
		}
		course.ThumbnailID = media.ID
	}

	// Handle introduction video
	if data.IntroductionVideoUrl != "" {
		if course.IntroductionVideoID != "" {
			if err := s.db.Model(&course.IntroductionVideo).Update("url", data.IntroductionVideoUrl).Error; err != nil {
				return err
			}
		} else {
			link := models.Link{URL: data.IntroductionVideoUrl}
			if err := s.db.Create(&link).Error; err != nil {
				return err
			}
			course.IntroductionVideoID = link.ID
		}
	} else if course.IntroductionVideoID != "" {
		if err := s.db.Delete(&models.Link{}, "id = ?", course.IntroductionVideoID).Error; err != nil {
			return err
		}
		course.IntroductionVideoID = ""
	}

	// Handle audience
	if len(data.Audiences) > 0 {
		audiences, err := s.resolveAudiences(data.Audiences)
		if err != nil {
			return err
		}
		if err := s.db.Model(&course).Association("Audiences").Replace(audiences); err != nil {
			return err
		}
	} else {
		if err := s.db.Model(&course).Association("Audiences").Clear(); err != nil {
			return err
		}
	}

	// Handle deleted course-level docs
	for _, mediaId := range data.DeletedDocIds {
		if err := s.db.Model(&course).Association("DownloadableContent").Unscoped().Delete(&models.Media{ID: mediaId}); err != nil {
			log.Printf("failed to unassociate doc %s: %v", mediaId, err)
		}
		if err := s.deleteMedia(mediaId); err != nil {
			log.Printf("failed to delete doc %s: %v", mediaId, err)
		}
	}

	// Upload new course-level docs
	if len(data.DocFiles) > 0 {
		medias, err := s.uploadMediaFilesWithNames(data.DocFiles, data.DocNames)
		if err != nil {
			return err
		}
		if err := s.db.Model(&course).Association("DownloadableContent").Append(medias); err != nil {
			return err
		}
	}

	if err := s.db.Save(&course).Error; err != nil {
		return err
	}

	// Handle chapters
	if err := s.syncChapters(&course, chapters, data.ChapterDocFiles, data.DeletedChapterIds); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCourse(id string) error {
	var course models.Course
	if err := s.db.Where("id = ?", id).
		Preload("Thumbnail").
		Preload("DownloadableContent").
		Preload("Chapters").
		Preload("Chapters.DownloadableContent").
		Preload("Chapters.VideoLink").
		First(&course).Error; err != nil {
		return err
	}

	// Delete chapters
	for _, ch := range course.Chapters {
		if err := s.deleteChapter(&ch); err != nil {
			log.Printf("failed to delete chapter %s: %v", ch.ID, err)
		}
	}

	// Delete course-level docs
	for _, m := range course.DownloadableContent {
		if err := s.deleteMedia(m.ID); err != nil {
			log.Printf("failed to delete doc %s: %v", m.ID, err)
		}
	}
	if err := s.db.Model(&course).Association("DownloadableContent").Clear(); err != nil {
		log.Printf("failed to clear downloadable content: %v", err)
	}

	// Delete thumbnail
	if course.ThumbnailID != "" {
		if err := s.deleteMedia(course.ThumbnailID); err != nil {
			log.Printf("failed to delete thumbnail %s: %v", course.ThumbnailID, err)
		}
	}

	if err := s.db.Model(&course).Association("Audiences").Clear(); err != nil {
		log.Printf("failed to clear audiences: %v", err)
	}

	return s.db.Delete(&course).Error
}

// createChapters creates new chapters for a course.
// chapterDocFiles is a flat array matched by newDocNames order across chapters.
func (s *Service) createChapters(courseID string, chapters []dto.ChapterInput, chapterDocFiles []*multipart.FileHeader) error {
	fileIdx := 0
	for _, ch := range chapters {
		chapter := models.CourseChapter{
			CourseID:    courseID,
			Title:       ch.Title,
			Description: ch.Description,
			IsFree:      ch.IsFree,
		}

		if ch.VideoUrl != "" {
			link := models.Link{URL: ch.VideoUrl}
			if err := s.db.Create(&link).Error; err != nil {
				return err
			}
			chapter.VideoLinkID = link.ID
		}

		if err := s.db.Create(&chapter).Error; err != nil {
			return err
		}

		// Upload docs for this chapter
		numNewDocs := len(ch.NewDocNames)
		if numNewDocs > 0 && fileIdx < len(chapterDocFiles) {
			end := fileIdx + numNewDocs
			if end > len(chapterDocFiles) {
				end = len(chapterDocFiles)
			}
			files := chapterDocFiles[fileIdx:end]
			names := ch.NewDocNames[:len(files)]
			medias, err := s.uploadMediaFilesWithNames(files, names)
			if err != nil {
				return err
			}
			if err := s.db.Model(&chapter).Association("DownloadableContent").Append(medias); err != nil {
				return err
			}
			fileIdx = end
		}
	}
	return nil
}

// syncChapters handles create/update/delete of chapters during course update.
func (s *Service) syncChapters(course *models.Course, chapters []dto.ChapterInput, chapterDocFiles []*multipart.FileHeader, deletedChapterIds []string) error {
	// Delete explicitly removed chapters
	for _, chId := range deletedChapterIds {
		var ch models.CourseChapter
		if err := s.db.Preload("DownloadableContent").Preload("VideoLink").First(&ch, "id = ?", chId).Error; err != nil {
			continue
		}
		if err := s.deleteChapter(&ch); err != nil {
			log.Printf("failed to delete chapter %s: %v", chId, err)
		}
	}

	fileIdx := 0
	for _, chInput := range chapters {
		numNewDocs := len(chInput.NewDocNames)

		var files []*multipart.FileHeader
		if numNewDocs > 0 && fileIdx < len(chapterDocFiles) {
			end := fileIdx + numNewDocs
			if end > len(chapterDocFiles) {
				end = len(chapterDocFiles)
			}
			files = chapterDocFiles[fileIdx:end]
			fileIdx = end
		}

		if chInput.ID == "" {
			// New chapter
			chapter := models.CourseChapter{
				CourseID:    course.ID,
				Title:       chInput.Title,
				Description: chInput.Description,
				IsFree:      chInput.IsFree,
			}
			if chInput.VideoUrl != "" {
				link := models.Link{URL: chInput.VideoUrl}
				if err := s.db.Create(&link).Error; err != nil {
					return err
				}
				chapter.VideoLinkID = link.ID
			}
			if err := s.db.Create(&chapter).Error; err != nil {
				return err
			}
			if len(files) > 0 {
				names := chInput.NewDocNames[:len(files)]
				medias, err := s.uploadMediaFilesWithNames(files, names)
				if err != nil {
					return err
				}
				if err := s.db.Model(&chapter).Association("DownloadableContent").Append(medias); err != nil {
					return err
				}
			}
		} else {
			// Update existing chapter
			var chapter models.CourseChapter
			if err := s.db.Preload("DownloadableContent").Preload("VideoLink").First(&chapter, "id = ?", chInput.ID).Error; err != nil {
				continue
			}
			chapter.Title = chInput.Title
			chapter.Description = chInput.Description
			chapter.IsFree = chInput.IsFree

			// Handle video link
			if chInput.VideoUrl != "" {
				if chapter.VideoLinkID != "" {
					if err := s.db.Model(&chapter.VideoLink).Update("url", chInput.VideoUrl).Error; err != nil {
						return err
					}
				} else {
					link := models.Link{URL: chInput.VideoUrl}
					if err := s.db.Create(&link).Error; err != nil {
						return err
					}
					chapter.VideoLinkID = link.ID
				}
			} else if chapter.VideoLinkID != "" {
				if err := s.db.Delete(&models.Link{}, "id = ?", chapter.VideoLinkID).Error; err != nil {
					return err
				}
				chapter.VideoLinkID = ""
			}

			// Delete removed docs
			for _, docId := range chInput.DeletedDocIds {
				if err := s.db.Model(&chapter).Association("DownloadableContent").Unscoped().Delete(&models.Media{ID: docId}); err != nil {
					log.Printf("failed to unassociate chapter doc %s: %v", docId, err)
				}
				if err := s.deleteMedia(docId); err != nil {
					log.Printf("failed to delete chapter doc %s: %v", docId, err)
				}
			}

			// Upload new docs
			if len(files) > 0 {
				names := chInput.NewDocNames[:len(files)]
				medias, err := s.uploadMediaFilesWithNames(files, names)
				if err != nil {
					return err
				}
				if err := s.db.Model(&chapter).Association("DownloadableContent").Append(medias); err != nil {
					return err
				}
			}

			if err := s.db.Save(&chapter).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) deleteChapter(ch *models.CourseChapter) error {
	for _, m := range ch.DownloadableContent {
		if err := s.deleteMedia(m.ID); err != nil {
			log.Printf("failed to delete chapter doc %s: %v", m.ID, err)
		}
	}
	if err := s.db.Model(ch).Association("DownloadableContent").Clear(); err != nil {
		log.Printf("failed to clear chapter downloadable content: %v", err)
	}
	if ch.VideoLinkID != "" {
		if err := s.db.Delete(&models.Link{}, "id = ?", ch.VideoLinkID).Error; err != nil {
			log.Printf("failed to delete chapter video link %s: %v", ch.VideoLinkID, err)
		}
	}
	return s.db.Delete(ch).Error
}

func (s *Service) deleteMedia(id string) error {
	if err := s.db.Delete(&models.Media{}, "id = ?", id).Error; err != nil {
		return err
	}
	return s.s3.Delete(id)
}

func (s *Service) resolveAudiences(names []string) ([]models.Audience, error) {
	var audiences []models.Audience
	if err := s.db.Where("name IN ?", names).Find(&audiences).Error; err != nil {
		return nil, err
	}
	return audiences, nil
}

func (s *Service) uploadSingleFile(file *multipart.FileHeader, name string) (models.Media, error) {
	f, err := file.Open()
	if err != nil {
		return models.Media{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	location, key, contentType, err := s.s3.UploadNetwork(f)
	if err != nil {
		return models.Media{}, err
	}
	return models.Media{ID: key, Name: name, URL: location, FileType: contentType}, nil
}

func (s *Service) uploadMediaFilesWithNames(files []*multipart.FileHeader, names []string) ([]models.Media, error) {
	medias := make([]models.Media, len(files))
	g, _ := errgroup.WithContext(context.Background())
	for i, file := range files {
		i, file := i, file
		name := ""
		if i < len(names) {
			name = names[i]
		}
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
			medias[i] = models.Media{ID: key, Name: name, URL: location, FileType: contentType}
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
