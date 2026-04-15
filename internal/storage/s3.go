package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3 struct {
	S3 *s3.Client
}

func InitS3() *S3 {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, option ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           os.Getenv("S3_URL"),
			SigningRegion: os.Getenv("S3_REGION"),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("S3_USERNAME"), os.Getenv("S3_PASSWORD"), os.Getenv("S3_SESSION"))))

	if err != nil {
		log.Fatal("Failed to initialize s3.")
		return nil
	}

	return &S3{
		S3: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
		})}
}

func (s *S3) UploadLocal(fileUrl string) (string, string, error) {
	uploader := manager.NewUploader(s.S3, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.LeavePartsOnError = false
	})
	file, err := os.Open(fileUrl)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	contentType := mime.TypeByExtension(filepath.Ext(fileUrl))
	if contentType == "" {
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		contentType = http.DetectContentType(buffer[:n])
	}
	file.Seek(0, 0)

	s3Key := uuid.Must(uuid.NewV7()).String()
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:             aws.String("expressions-india"),
		Key:                aws.String(s3Key),
		Body:               file,
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String("inline"),
	})
	if err != nil {
		return "", "", err
	}
	return result.Location, s3Key, nil
}
func (s *S3) UploadNetwork(file io.Reader) (string, string, string, error) {

	uploader := manager.NewUploader(s.S3)

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", "", "", fmt.Errorf("failed to read file header: %w", err)
	}
	contentType := http.DetectContentType(buffer[:n])

	fullBody := io.MultiReader(bytes.NewReader(buffer[:n]), file)

	s3Key := uuid.Must(uuid.NewV7()).String()
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String("expressions-india"),
		Key:         aws.String(s3Key),
		Body:        fullBody,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", "", "", fmt.Errorf("s3 upload failed: %w", err)
	}

	return result.Location, s3Key, contentType, nil
}

func (s *S3) DeleteFromS3(s3Key string) error {

	_, err := s.S3.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String("expressions-india"),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3) Delete(s3Key string) error {
	if err := s.DeleteFromS3(s3Key); err != nil {
		return err
	}
	return nil
}

// PresignUpload generates a presigned PUT URL for the given object key and
// content type. The URL expires after ttl and the client must supply a
// matching Content-Type header when uploading.
func (s *S3) PresignUpload(id, contentType string, ttl time.Duration) (presignedURL string, err error) {
	pc := s3.NewPresignClient(s.S3)
	req, err := pc.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String("expressions-india"),
		Key:         aws.String(id),
		ContentType: aws.String(contentType),
	}, func(o *s3.PresignOptions) {
		o.Expires = ttl
	})
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}
	return req.URL, nil
}

// PublicURL returns the publicly accessible URL for an already-uploaded object.
func (s *S3) PublicURL(id string) string {
	return fmt.Sprintf("%s/expressions-india/%s", os.Getenv("S3_URL"), id)
}
