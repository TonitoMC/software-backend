package exam

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"software-backend/internal/models"
	repository "software-backend/internal/repository/exam"
	s3service "software-backend/internal/service/s3"
)

type ExamService interface {
	GetByPatientID(patientID int) ([]models.Exam, error)
	UploadPDF(examID int, file *multipart.FileHeader) error
	GetDownloadURL(examID int) (string, error)
	StreamFile(examID int) (io.ReadCloser, *FileMetadata, error) // Add this line
	GetPending() ([]*models.Exam, error)
}

// Add this struct after the interface
type FileMetadata struct {
	FileName string
	MimeType string
	FileSize int64
}
type examService struct {
	repo      repository.ExamRepository
	s3Service s3service.S3Service
}

func NewExamService(repo repository.ExamRepository, s3Service s3service.S3Service) ExamService {
	return &examService{
		repo:      repo,
		s3Service: s3Service,
	}
}

func (s *examService) GetByPatientID(patientID int) ([]models.Exam, error) {
	return s.repo.GetByPatientID(patientID)
}

func (s *examService) UploadPDF(examID int, file *multipart.FileHeader) error {
	// Validate file
	if file.Size > 10*1024*1024 { // 10MB limit
		return fmt.Errorf("file too large (max 10MB)")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		return fmt.Errorf("only PDF files are allowed")
	}

	// Generate S3 key
	key := fmt.Sprintf("examenes/%d/%s_%s",
		examID,
		time.Now().Format("20060102_150405"),
		file.Filename)

	// Upload to S3
	err := s.s3Service.UploadFile(key, file)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Update database
	err = s.repo.UpdateFileMetadata(examID, key, file.Size, "application/pdf")
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	return nil
}

func (s *examService) StreamFile(examID int) (io.ReadCloser, *FileMetadata, error) {
	exam, err := s.repo.GetByID(examID)
	if err != nil {
		return nil, nil, fmt.Errorf("exam not found: %w", err)
	}

	if !exam.S3Key.Valid || exam.S3Key.String == "" {
		return nil, nil, fmt.Errorf("no file uploaded for this exam")
	}

	reader, err := s.s3Service.GetObject(exam.S3Key.String)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch file: %w", err)
	}

	metadata := &FileMetadata{
		FileName: filepath.Base(exam.S3Key.String),
		MimeType: "application/pdf",
		FileSize: exam.FileSize.Int64,
	}

	return reader, metadata, nil
}

func (s *examService) GetDownloadURL(examID int) (string, error) {
	exam, err := s.repo.GetByID(examID)
	if err != nil {
		return "", fmt.Errorf("exam not found: %w", err)
	}

	if !exam.S3Key.Valid || exam.S3Key.String == "" {
		return "", fmt.Errorf("no file uploaded for this exam")
	}

	url, err := s.s3Service.GeneratePresignedURL(exam.S3Key.String, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

func (s *examService) GetPending() ([]*models.Exam, error) {
	return s.repo.GetPending()
}
