package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ls1intum/prompt2/servers/certificate/config"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
	"github.com/ls1intum/prompt2/servers/certificate/participants"
	log "github.com/sirupsen/logrus"
)

type GeneratorService struct {
	queries db.Queries
}

var GeneratorServiceSingleton *GeneratorService

func NewGeneratorService(queries db.Queries) *GeneratorService {
	return &GeneratorService{
		queries: queries,
	}
}

type CertificateData struct {
	StudentName string `json:"studentName"`
	CourseName  string `json:"courseName"`
	TeamName    string `json:"teamName"`
	Date        string `json:"date"`
}

// writeDataFiles writes the certificate data JSON to both data.json and vars.json
// in the given directory, so templates using either filename convention will work.
func writeDataFiles(tempDir string, certData CertificateData) error {
	dataJSON, err := json.Marshal(certData)
	if err != nil {
		return fmt.Errorf("failed to marshal certificate data: %w", err)
	}
	for _, name := range []string{"data.json", "vars.json"} {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, dataJSON, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", name, err)
		}
	}
	return nil
}

// TypstCompilationError represents a Typst template compilation failure
// with the raw compiler output for display to the user.
type TypstCompilationError struct {
	Output string
}

func (e *TypstCompilationError) Error() string {
	return fmt.Sprintf("typst compilation failed: %s", e.Output)
}

// compileTypst runs the typst compiler on the template and returns the generated PDF bytes.
// A 30-second timeout is applied to prevent hung or infinite-loop Typst invocations.
func compileTypst(ctx context.Context, tempDir, templatePath string) ([]byte, error) {
	compilationTimeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(ctx, compilationTimeout)
	defer cancel()

	outputPath := filepath.Join(tempDir, "certificate.pdf")
	cmd := exec.CommandContext(ctx, "typst", "compile", templatePath, outputPath)
	cmd.Dir = tempDir

	if out, err := cmd.CombinedOutput(); err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"output": string(out),
		}).Error("Typst compilation failed")
		return nil, &TypstCompilationError{Output: string(out)}
	}

	pdfData, err := os.ReadFile(outputPath)
	if err != nil {
		log.WithError(err).Error("Failed to read generated PDF")
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}
	return pdfData, nil
}

func (s *GeneratorService) GenerateCertificate(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, student *participants.Student) ([]byte, error) {
	// Get template content
	templateContent, err := config.GetTemplateContent(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get template content")
		return nil, fmt.Errorf("no template configured: %w", err)
	}

	// Get course info
	coursePhase, err := participants.GetCoursePhaseWithCourse(ctx, authHeader, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get course info")
		return nil, fmt.Errorf("failed to get course info: %w", err)
	}

	// Resolve team name (best-effort, non-fatal if not available)
	teamName, err := participants.GetStudentTeamName(ctx, authHeader, coursePhaseID, student.ID)
	if err != nil {
		log.WithError(err).Warn("Could not resolve team name, continuing without it")
	}

	// Determine certificate date: use the configured release date for consistency,
	// so every download shows the same date. Fall back to today if not set.
	certDate := time.Now().Format("January 2, 2006")
	if phaseConfig, configErr := config.GetCoursePhaseConfig(ctx, coursePhaseID); configErr == nil && phaseConfig.ReleaseDate != nil {
		certDate = phaseConfig.ReleaseDate.Format("January 2, 2006")
	}

	// Create temp directory for processing
	tempDir, err := os.MkdirTemp("", "certificate-*")
	if err != nil {
		log.WithError(err).Error("Failed to create temp directory")
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Write template to temp file
	templatePath := filepath.Join(tempDir, "template.typ")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		log.WithError(err).Error("Failed to write template file")
		return nil, fmt.Errorf("failed to write template file: %w", err)
	}

	// Create certificate data
	certData := CertificateData{
		StudentName: fmt.Sprintf("%s %s", student.FirstName, student.LastName),
		CourseName:  coursePhase.Course.Name,
		TeamName:    teamName,
		Date:        certDate,
	}

	// Write data JSON files (data.json + vars.json for template compatibility)
	if err := writeDataFiles(tempDir, certData); err != nil {
		log.WithError(err).Error("Failed to write data files")
		return nil, err
	}

	// Compile template to PDF
	pdfData, err := compileTypst(ctx, tempDir, templatePath)
	if err != nil {
		return nil, err
	}

	return pdfData, nil
}

// GeneratePreviewCertificate creates a certificate PDF using mock data and the saved template.
// This is used by instructors to preview how the certificate will look.
func (s *GeneratorService) GeneratePreviewCertificate(ctx context.Context, coursePhaseID uuid.UUID) ([]byte, error) {
	// Get template content
	templateContent, err := config.GetTemplateContent(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get template content")
		return nil, fmt.Errorf("no template configured: %w", err)
	}

	// Create temp directory for processing
	tempDir, err := os.MkdirTemp("", "certificate-preview-*")
	if err != nil {
		log.WithError(err).Error("Failed to create temp directory")
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Write template to temp file
	templatePath := filepath.Join(tempDir, "template.typ")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		log.WithError(err).Error("Failed to write template file")
		return nil, fmt.Errorf("failed to write template file: %w", err)
	}

	// Create mock certificate data
	certData := CertificateData{
		StudentName: "Jane Doe",
		CourseName:  "iPraktikum",
		TeamName:    "Hermes",
		Date:        time.Now().Format("January 2, 2006"),
	}

	// Write data JSON files (data.json + vars.json for template compatibility)
	if err := writeDataFiles(tempDir, certData); err != nil {
		log.WithError(err).Error("Failed to write data files")
		return nil, err
	}

	// Compile template to PDF
	return compileTypst(ctx, tempDir, templatePath)
}
