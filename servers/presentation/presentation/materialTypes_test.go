package presentation

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeMaterialTypesOrdersAndDeduplicates(t *testing.T) {
	normalized, err := normalizeMaterialTypes([]string{
		MaterialTypeRecording, MaterialTypeSlides, " slides ", "", MaterialTypeHandout,
	})
	require.NoError(t, err)
	assert.Equal(t,
		[]string{MaterialTypeSlides, MaterialTypeHandout, MaterialTypeRecording},
		normalized,
		"requirements are stored in catalog order so the stored list is independent of request order")
}

func TestNormalizeMaterialTypesRejectsUnknownKeys(t *testing.T) {
	_, err := normalizeMaterialTypes([]string{MaterialTypeSlides, "whitepaper"})

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 400, apiErr.Status)
	assert.Equal(t, "invalid_material_type", apiErr.Code)
}

func TestNormalizeMaterialTypesAcceptsEmptySelection(t *testing.T) {
	normalized, err := normalizeMaterialTypes(nil)
	require.NoError(t, err)
	assert.Empty(t, normalized, "a phase may ask for no uploads at all")
}

// The deployment-wide allow-list is the outer gate, so it has to cover every media type a
// lecturer is able to require. Otherwise enabling a slot would produce uploads that always
// fail with material_type_not_allowed.
func TestDefaultAllowedTypesCoverTheWholeCatalog(t *testing.T) {
	allowed := DefaultAllowedMaterialTypes()

	for _, definition := range materialTypeCatalog {
		for _, contentType := range definition.ContentTypes {
			assert.True(t, slices.Contains(allowed, contentType),
				"%s is required for %s uploads but not allowed by default", contentType, definition.Type)
		}
	}
}

func TestEnsureMaterialTypeRequestedRejectsUnaskedType(t *testing.T) {
	err := ensureMaterialTypeRequested([]string{MaterialTypeSlides}, MaterialTypeCode)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 409, apiErr.Status)
	assert.Equal(t, "material_type_not_requested", apiErr.Code)
}

// A presenter picking exactly the requested file must not be blocked because their machine
// has no handler registered for the extension and the browser sent nothing.
func TestResolveContentTypeFallsBackToTheExtension(t *testing.T) {
	service := &Service{allowedTypes: DefaultAllowedMaterialTypes()}

	resolved := resolveContentType("Final deck.odp", "")
	assert.Equal(t, "application/vnd.oasis.opendocument.presentation", resolved)
	require.NoError(t, service.ensureContentTypeForMaterial(MaterialTypeSlides, resolved))

	resolved = resolveContentType("REPORT.ODT", unknownContentType)
	assert.Equal(t, "application/vnd.oasis.opendocument.text", resolved)
	require.NoError(t, service.ensureContentTypeForMaterial(MaterialTypeSummary, resolved))

	// A browser that knows the type keeps the last word, so a renamed file is still caught.
	assert.Equal(t, "video/mp4", resolveContentType("clip.pdf", "video/mp4"))

	// Nothing to go on stays unknown and is rejected as before.
	resolved = resolveContentType("archive.rar", "")
	assert.Equal(t, unknownContentType, resolved)
	assert.Error(t, service.ensureContentTypeForMaterial(MaterialTypeCode, resolved))
}

// The fallback may only produce media types the catalog actually accepts, otherwise it
// swaps one hard rejection for another.
func TestExtensionFallbackStaysInsideTheCatalog(t *testing.T) {
	allowed := DefaultAllowedMaterialTypes()

	for extension, contentType := range mediaTypeByExtension {
		assert.True(t, slices.Contains(allowed, contentType),
			"%s resolves to %s, which no material type accepts", extension, contentType)
	}
}

func TestEnsureContentTypeForMaterialRejectsMismatch(t *testing.T) {
	service := &Service{allowedTypes: DefaultAllowedMaterialTypes()}

	require.NoError(t, service.ensureContentTypeForMaterial(MaterialTypeCode, "application/zip"))
	require.NoError(t, service.ensureContentTypeForMaterial(
		MaterialTypeSlides, "application/pdf; charset=binary"))

	// Globally allowed, but not a slide deck.
	err := service.ensureContentTypeForMaterial(MaterialTypeSlides, "video/mp4")
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 415, apiErr.Status)
	assert.Equal(t, "material_type_mismatch", apiErr.Code)
}
