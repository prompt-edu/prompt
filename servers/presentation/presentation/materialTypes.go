package presentation

import (
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

const unknownContentType = "application/octet-stream"

// The material types students can be asked to upload. These keys are part of the API and
// are mirrored by the client's label catalog in
// clients/presentation_component/src/presentation/materialTypes.ts, so the list may grow
// but a key must never be renamed.
const (
	MaterialTypeSlides    = "slides"
	MaterialTypeSummary   = "summary"
	MaterialTypeHandout   = "handout"
	MaterialTypePoster    = "poster"
	MaterialTypeCode      = "code"
	MaterialTypeRecording = "recording"
)

type materialTypeDefinition struct {
	Type string
	// The media types a file of this material type may have. The server is the authority
	// here; the client only uses its own extension list to prefill the file dialog.
	ContentTypes []string
}

// Ordered: normalizeMaterialTypes stores a phase's requirements in this order, which is
// also the order the client renders the upload slots in.
var materialTypeCatalog = []materialTypeDefinition{
	{
		Type: MaterialTypeSlides,
		ContentTypes: []string{
			"application/pdf",
			"application/vnd.ms-powerpoint",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"application/vnd.oasis.opendocument.presentation",
		},
	},
	{
		Type: MaterialTypeSummary,
		ContentTypes: []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.oasis.opendocument.text",
		},
	},
	{
		Type:         MaterialTypeHandout,
		ContentTypes: []string{"application/pdf"},
	},
	{
		Type:         MaterialTypePoster,
		ContentTypes: []string{"application/pdf", "image/png", "image/jpeg"},
	},
	{
		Type:         MaterialTypeCode,
		ContentTypes: []string{"application/zip", "application/x-zip-compressed"},
	},
	{
		Type:         MaterialTypeRecording,
		ContentTypes: []string{"video/mp4"},
	},
}

// A phase that was never configured still has to accept the slide deck every presentation
// needs, so the schema default and this fallback agree on it.
var defaultRequiredMaterialTypes = []string{MaterialTypeSlides}

// The file endings the catalog's media types are normally stored under, mirroring the
// client's extension lists. Every value must appear in materialTypeCatalog.
var mediaTypeByExtension = map[string]string{
	".pdf":  "application/pdf",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".odp":  "application/vnd.oasis.opendocument.presentation",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".odt":  "application/vnd.oasis.opendocument.text",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".zip":  "application/zip",
	".mp4":  "video/mp4",
}

// Browsers send an empty file type whenever the machine has no handler registered for the
// extension, which is the normal case for .odp or .odt without an office suite installed.
// The file name is then the only signal left, and both are equally client-controlled.
func resolveContentType(fileName, contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType != "" && !strings.EqualFold(mediaType(contentType), unknownContentType) {
		return contentType
	}
	if resolved, exists := mediaTypeByExtension[strings.ToLower(filepath.Ext(fileName))]; exists {
		return resolved
	}
	if contentType == "" {
		return unknownContentType
	}
	return contentType
}

// DefaultAllowedMaterialTypes is the union of every catalog entry's media types. It is the
// default for PRESENTATION_ALLOWED_FILE_TYPES, so the deployment-wide allow-list cannot
// silently exclude a type a lecturer is able to require.
func DefaultAllowedMaterialTypes() []string {
	allowed := make([]string, 0)
	for _, definition := range materialTypeCatalog {
		for _, contentType := range definition.ContentTypes {
			if !slices.Contains(allowed, contentType) {
				allowed = append(allowed, contentType)
			}
		}
	}
	return allowed
}

func materialTypeContentTypes(materialType string) ([]string, bool) {
	for _, definition := range materialTypeCatalog {
		if definition.Type == materialType {
			return definition.ContentTypes, true
		}
	}
	return nil, false
}

// normalizeMaterialTypes rejects unknown keys and returns the requested types deduplicated
// and in catalog order, so a stored requirement list never depends on request order.
func normalizeMaterialTypes(values []string) ([]string, error) {
	requested := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := materialTypeContentTypes(trimmed); !exists {
			return nil, apiError(http.StatusBadRequest, "invalid_material_type",
				fmt.Sprintf("Material type %q does not exist", trimmed), nil)
		}
		if !slices.Contains(requested, trimmed) {
			requested = append(requested, trimmed)
		}
	}
	normalized := make([]string, 0, len(requested))
	for _, definition := range materialTypeCatalog {
		if slices.Contains(requested, definition.Type) {
			normalized = append(normalized, definition.Type)
		}
	}
	return normalized, nil
}

// A file may only be uploaded for a material type the phase actually asks for, which also
// keeps a stale client from filling a slot the lecturer has since removed.
func ensureMaterialTypeRequested(requiredTypes []string, materialType string) error {
	if slices.Contains(requiredTypes, materialType) {
		return nil
	}
	return apiError(http.StatusConflict, "material_type_not_requested",
		fmt.Sprintf("This phase does not ask for %q uploads", materialType), nil)
}

// ensureContentTypeForMaterial applies both file type gates: the deployment-wide allow-list
// and the media types the material type itself permits. Completion re-runs this against the
// stored object, so it deliberately does not care whether the phase still asks for the type.
func (s *Service) ensureContentTypeForMaterial(materialType, contentType string) error {
	if err := s.ensureAllowedType(contentType); err != nil {
		return err
	}
	contentTypes, exists := materialTypeContentTypes(materialType)
	if !exists {
		return apiError(http.StatusBadRequest, "invalid_material_type",
			fmt.Sprintf("Material type %q does not exist", materialType), nil)
	}
	if slices.ContainsFunc(contentTypes, func(allowed string) bool {
		return strings.EqualFold(mediaType(allowed), mediaType(contentType))
	}) {
		return nil
	}
	return apiError(http.StatusUnsupportedMediaType, "material_type_mismatch",
		fmt.Sprintf("A %q upload cannot have the file type %q", materialType, contentType), nil)
}
