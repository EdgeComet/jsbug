package server

import (
	"testing"

	"github.com/user/jsbug/internal/types"
)

func TestEnrichImagesWithSizes_Basic(t *testing.T) {
	images := []types.Image{
		{Src: "https://example.com/image1.jpg"},
		{Src: "https://example.com/image2.png"},
	}

	requests := []types.NetworkRequest{
		{URL: "https://example.com/image1.jpg", Type: "Image", Size: 1024},
		{URL: "https://example.com/image2.png", Type: "Image", Size: 2048},
	}

	enrichImagesWithSizes(images, requests)

	if images[0].Size != 1024 {
		t.Errorf("Image[0].Size = %d, want 1024", images[0].Size)
	}
	if images[1].Size != 2048 {
		t.Errorf("Image[1].Size = %d, want 2048", images[1].Size)
	}
}

func TestEnrichImagesWithSizes_NoMatch(t *testing.T) {
	images := []types.Image{
		{Src: "https://example.com/image.jpg"},
	}

	requests := []types.NetworkRequest{
		{URL: "https://other.com/image.jpg", Type: "Image", Size: 1024},
	}

	enrichImagesWithSizes(images, requests)

	// Size should remain 0 when no match
	if images[0].Size != 0 {
		t.Errorf("Image[0].Size = %d, want 0 (no match)", images[0].Size)
	}
}

func TestEnrichImagesWithSizes_OnlyImageType(t *testing.T) {
	images := []types.Image{
		{Src: "https://example.com/script.js"},
		{Src: "https://example.com/image.jpg"},
	}

	requests := []types.NetworkRequest{
		{URL: "https://example.com/script.js", Type: "Script", Size: 5000},
		{URL: "https://example.com/image.jpg", Type: "Image", Size: 1024},
	}

	enrichImagesWithSizes(images, requests)

	// Script type should not match
	if images[0].Size != 0 {
		t.Errorf("Image[0].Size = %d, want 0 (wrong type)", images[0].Size)
	}
	// Image type should match
	if images[1].Size != 1024 {
		t.Errorf("Image[1].Size = %d, want 1024", images[1].Size)
	}
}

func TestEnrichImagesWithSizes_SkipZeroSize(t *testing.T) {
	images := []types.Image{
		{Src: "https://example.com/image.jpg"},
	}

	requests := []types.NetworkRequest{
		{URL: "https://example.com/image.jpg", Type: "Image", Size: 0},
	}

	enrichImagesWithSizes(images, requests)

	// Should not match requests with Size == 0
	if images[0].Size != 0 {
		t.Errorf("Image[0].Size = %d, want 0", images[0].Size)
	}
}

func TestEnrichImagesWithSizes_MixedMatches(t *testing.T) {
	images := []types.Image{
		{Src: "https://example.com/found.jpg"},
		{Src: "https://example.com/notfound.jpg"},
		{Src: "https://cdn.example.com/external.png"},
	}

	requests := []types.NetworkRequest{
		{URL: "https://example.com/found.jpg", Type: "Image", Size: 1000},
		{URL: "https://cdn.example.com/external.png", Type: "Image", Size: 3000},
		{URL: "https://example.com/other.jpg", Type: "Image", Size: 500},
	}

	enrichImagesWithSizes(images, requests)

	if images[0].Size != 1000 {
		t.Errorf("Image[0].Size = %d, want 1000", images[0].Size)
	}
	if images[1].Size != 0 {
		t.Errorf("Image[1].Size = %d, want 0 (not found)", images[1].Size)
	}
	if images[2].Size != 3000 {
		t.Errorf("Image[2].Size = %d, want 3000", images[2].Size)
	}
}

func TestEnrichImagesWithSizes_EmptyInputs(t *testing.T) {
	// Empty images
	var images []types.Image
	requests := []types.NetworkRequest{
		{URL: "https://example.com/image.jpg", Type: "Image", Size: 1024},
	}
	enrichImagesWithSizes(images, requests) // Should not panic

	// Empty requests
	images = []types.Image{
		{Src: "https://example.com/image.jpg"},
	}
	enrichImagesWithSizes(images, nil)

	if images[0].Size != 0 {
		t.Errorf("Image[0].Size = %d, want 0", images[0].Size)
	}
}

func TestEnrichImagesWithSizes_PreservesOtherFields(t *testing.T) {
	images := []types.Image{
		{
			Src:        "https://example.com/image.jpg",
			Alt:        "Test image",
			IsExternal: false,
			IsAbsolute: true,
			IsInLink:   true,
			LinkHref:   "/gallery",
		},
	}

	requests := []types.NetworkRequest{
		{URL: "https://example.com/image.jpg", Type: "Image", Size: 2048},
	}

	enrichImagesWithSizes(images, requests)

	// Size should be set
	if images[0].Size != 2048 {
		t.Errorf("Size = %d, want 2048", images[0].Size)
	}

	// Other fields should be preserved
	if images[0].Alt != "Test image" {
		t.Errorf("Alt = %q, want %q", images[0].Alt, "Test image")
	}
	if images[0].IsExternal != false {
		t.Error("IsExternal should be false")
	}
	if images[0].IsAbsolute != true {
		t.Error("IsAbsolute should be true")
	}
	if images[0].IsInLink != true {
		t.Error("IsInLink should be true")
	}
	if images[0].LinkHref != "/gallery" {
		t.Errorf("LinkHref = %q, want %q", images[0].LinkHref, "/gallery")
	}
}
