package dto

import (
	"encoding/json"
	"time"
)

type CourseListItemDTO struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	Audiences    []string  `json:"audiences"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CourseMediaDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	FileType string `json:"fileType"`
}

// CourseChapterSummaryDTO is returned as part of the course overview — no paid content.
type CourseChapterSummaryDTO struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	IsFree bool   `json:"isFree"`
}

// CourseChapterDTO is returned by the single-chapter endpoint — full paid content.
type CourseChapterDTO struct {
	ID                  string           `json:"id"`
	Title               string           `json:"title"`
	Description         string           `json:"description"`
	VideoLinkURL        string           `json:"videoLinkUrl"`
	IsFree              bool             `json:"isFree"`
	DownloadableContent []CourseMediaDTO `json:"downloadableContent"`
}

type CourseDTO struct {
	ID                   string                    `json:"id"`
	Title                string                    `json:"title"`
	Description          string                    `json:"description"`
	Thumbnail            *CourseMediaDTO           `json:"thumbnail"`
	IntroductionVideoURL string                    `json:"introductionVideoUrl"`
	RegistrationURL      string                    `json:"registrationUrl"`
	DownloadableContent  []CourseMediaDTO          `json:"downloadableContent"`
	Audiences            []string                  `json:"audiences"`
	Chapters             []CourseChapterSummaryDTO `json:"chapters"`
	CreatedAt            time.Time                 `json:"createdAt"`
	UpdatedAt            time.Time                 `json:"updatedAt"`
}

// ChapterInput is parsed from the chaptersJson form field.
type ChapterInput struct {
	ID             string             `json:"id"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	VideoUrl       string             `json:"videoUrl"`
	IsFree         bool               `json:"isFree"`
	NewDocUploads  []UploadedMediaRef `json:"newDocUploads"`  // pre-uploaded docs for this chapter
	ExistingDocIds []string           `json:"existingDocIds"` // IDs of existing docs to keep
	DeletedDocIds  []string           `json:"deletedDocIds"`  // IDs of existing docs to delete (update only)
}

type CourseCreateRequestDTO struct {
	Title                string   `form:"title" binding:"required"`
	Description          string   `form:"description"`
	Audiences            []string `form:"audiences"`
	IntroductionVideoUrl string   `form:"introductionVideoUrl"`
	RegistrationURL      string   `form:"registrationUrl"`
	// Pre-uploaded thumbnail (JSON-encoded UploadedMediaRef). Empty = no thumbnail.
	ThumbnailUpload    string   `form:"thumbnailUpload"`
	DeletedThumbnailId string   `form:"deletedThumbnailId"`
	// Course-level downloadable content: each element is a JSON-encoded UploadedMediaRef.
	DocUploads    []string `form:"docUploads"`
	DeletedDocIds []string `form:"deletedDocIds"`
	// Chapters as JSON ([]ChapterInput). Chapter doc files are embedded in NewDocUploads.
	ChaptersJson      string   `form:"chaptersJson"`
	DeletedChapterIds []string `form:"deletedChapterIds"`
}

type EnrolledUserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type EnrollUserRequest struct {
	UserID string `json:"userId" binding:"required"`
}

func (r *CourseCreateRequestDTO) ParsedChapters() ([]ChapterInput, error) {
	if r.ChaptersJson == "" {
		return nil, nil
	}
	var chapters []ChapterInput
	if err := json.Unmarshal([]byte(r.ChaptersJson), &chapters); err != nil {
		return nil, err
	}
	return chapters, nil
}
