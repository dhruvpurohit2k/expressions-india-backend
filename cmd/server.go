package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type Server struct {
	mux *http.ServeMux
	mu  sync.RWMutex
	db  *sqlx.DB
	s3  *s3.Client
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) SetupRoutes() {
	s.mux.HandleFunc("GET /event", s.GetEventList)
	s.mux.HandleFunc("GET /event/{id}", s.GetEvent)
	s.mux.HandleFunc("POST /event", s.HandleCreateEvent)
	s.mux.HandleFunc("PUT /event/{id}", s.UpdateEvent)
	s.mux.HandleFunc("GET /workshop", s.GetWorkshops)
	s.mux.HandleFunc("POST /workshop", s.PostWorkshop)
}

func createServer() *Server {
	server := &Server{
		mux: &http.ServeMux{},
		mu:  sync.RWMutex{},
	}

	dbConn, err := connectToDB()
	if err != nil {
		return nil
	}
	server.db = dbConn

	s3Conn, err := connectToS3()
	if err != nil {
		return nil
	}
	server.s3 = s3Conn

	if "true" == os.Getenv("SEEDING_DB") {
		if err := seedDB(server); err != nil {
			fmt.Println(err)
			return nil
		}
	}

	server.SetupRoutes()
	return server
}

func connectToDB() (*sqlx.DB, error) {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Could not load ENV", err)
		return nil, err
	}
	dbString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("pgx", dbString)
	if err != nil {
		fmt.Println("Could not open DB", err.Error())
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(0)
	if err = runMigrations(dbString); err != nil {
		fmt.Println("ERROR WHILE RUNNING MIGRATION")
		return nil, err
	}

	return sqlx.NewDb(db, "pgx"), nil

}

func connectToS3() (*s3.Client, error) {
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
		return nil, err
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func runMigrations(dbString string) error {
	m, err := migrate.New(os.Getenv("MIGRATION_PATH"), dbString)
	if err != nil {
		fmt.Println("MIGRATION FAILED")
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	} else {
		return nil
	}
}

func (s *Server) uploadTos3(fileUrl, fileName string, folderName string) (string, string, error) {
	uploader := manager.NewUploader(s.s3, func(u *manager.Uploader) {
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

	var s3Key string
	hash := sha256.Sum256([]byte(fileName))
	hashString := hex.EncodeToString(hash[:])
	if folderName != "" {
		s3Key = fmt.Sprintf("%s/%s", folderName, hashString)
	} else {
		s3Key = uuid.NewString()
	}
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:             aws.String("expressions-india"),
		Key:                aws.String(s3Key),
		Body:               file,
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String("inline"),
	})
	return result.Location, s3Key, nil
}
func (s *Server) uploadTos3IO(file io.Reader, fileName, folderName string) (string, string, error) {

	uploader := manager.NewUploader(s.s3)

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", "", fmt.Errorf("failed to read file header: %w", err)
	}
	contentType := http.DetectContentType(buffer[:n])

	fullBody := io.MultiReader(bytes.NewReader(buffer[:n]), file)

	ext := filepath.Ext(fileName)
	stringToHash := fmt.Sprintf("%s/%s", folderName, fileName)
	hashBytes := sha256.Sum256([]byte(stringToHash))
	hashString := hex.EncodeToString(hashBytes[:])

	s3Key := fmt.Sprintf("%s/%s%s", folderName, hashString, ext)

	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String("expressions-india"),
		Key:         aws.String(s3Key),
		Body:        fullBody,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", "", fmt.Errorf("s3 upload failed: %w", err)
	}

	return result.Location, s3Key, nil
}

func (s *Server) DeleteFromS3(s3Key string) error {

	_, err := s.s3.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String("expressions-india"),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return err
	}
	return nil
}
