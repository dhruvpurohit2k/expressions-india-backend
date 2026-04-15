package dto

// PresignRequestItem describes a single file the client wants to upload.
type PresignRequestItem struct {
	ContentType string `json:"contentType" binding:"required"`
	FileName    string `json:"fileName"`
}

// PresignResponseItem is returned for each requested file.
// The client must PUT the file bytes to PresignedURL with the matching
// Content-Type header, then include ID/Name/FileType in the final form submit.
type PresignResponseItem struct {
	ID           string `json:"id"`
	PresignedURL string `json:"presignedUrl"`
	URL          string `json:"url"`
	ContentType  string `json:"contentType"`
	FileName     string `json:"fileName"`
}

// UploadedMediaRef is passed (JSON-encoded) in form fields for each file that
// was already uploaded directly to S3 via a presigned URL.
type UploadedMediaRef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FileType string `json:"fileType"`
}
