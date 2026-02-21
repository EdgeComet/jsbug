package compare

import (
	"testing"

	"github.com/user/jsbug/internal/types"
)

func TestComputeRenderingImpact_None(t *testing.T) {
	jsData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	nonJSData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	diff := &types.CompareDiff{}

	impact := ComputeRenderingImpact(jsData, nonJSData, diff)

	if impact.OverallChange != "none" {
		t.Errorf("expected overall_change='none', got %q", impact.OverallChange)
	}
	if impact.TitleChanged {
		t.Error("expected TitleChanged=false")
	}
	if impact.MetaDescChanged {
		t.Error("expected MetaDescChanged=false")
	}
	if impact.CanonicalChanged {
		t.Error("expected CanonicalChanged=false")
	}
	if impact.H1Changed {
		t.Error("expected H1Changed=false")
	}
	if impact.ContentChangePercent != 0 {
		t.Errorf("expected ContentChangePercent=0, got %f", impact.ContentChangePercent)
	}
}

func TestComputeRenderingImpact_Minor(t *testing.T) {
	jsData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       106,
	}
	nonJSData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	diff := &types.CompareDiff{}

	impact := ComputeRenderingImpact(jsData, nonJSData, diff)

	if impact.OverallChange != "minor" {
		t.Errorf("expected overall_change='minor', got %q", impact.OverallChange)
	}
	// 6/106*100 = 5.660... rounded to 5.7
	if impact.ContentChangePercent != 5.7 {
		t.Errorf("expected ContentChangePercent=5.7, got %f", impact.ContentChangePercent)
	}
}

func TestComputeRenderingImpact_Major_TitleChanged(t *testing.T) {
	jsData := &types.RenderData{
		Title:           "JS Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	nonJSData := &types.RenderData{
		Title:           "Different Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	diff := &types.CompareDiff{}

	impact := ComputeRenderingImpact(jsData, nonJSData, diff)

	if impact.OverallChange != "major" {
		t.Errorf("expected overall_change='major', got %q", impact.OverallChange)
	}
	if !impact.TitleChanged {
		t.Error("expected TitleChanged=true")
	}
}

func TestComputeRenderingImpact_Major_HighContentChange(t *testing.T) {
	jsData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       400,
	}
	nonJSData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	diff := &types.CompareDiff{}

	impact := ComputeRenderingImpact(jsData, nonJSData, diff)

	if impact.OverallChange != "major" {
		t.Errorf("expected overall_change='major', got %q", impact.OverallChange)
	}
	// 300/400*100 = 75.0
	if impact.ContentChangePercent != 75.0 {
		t.Errorf("expected ContentChangePercent=75.0, got %f", impact.ContentChangePercent)
	}
}

func TestComputeRenderingImpact_Major_ManyLinksAdded(t *testing.T) {
	jsData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	nonJSData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}

	addedLinks := make([]types.Link, 11)
	for i := range addedLinks {
		addedLinks[i] = types.Link{Href: "https://example.com/link"}
	}
	diff := &types.CompareDiff{
		Links: &types.LinksDiff{
			Added:   addedLinks,
			Removed: []types.Link{},
		},
	}

	impact := ComputeRenderingImpact(jsData, nonJSData, diff)

	if impact.OverallChange != "major" {
		t.Errorf("expected overall_change='major', got %q", impact.OverallChange)
	}
	if impact.LinksAdded != 11 {
		t.Errorf("expected LinksAdded=11, got %d", impact.LinksAdded)
	}
}

func TestComputeRenderingImpact_Major_ManyImagesAdded(t *testing.T) {
	jsData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}
	nonJSData := &types.RenderData{
		Title:           "Same Title",
		MetaDescription: "Same desc",
		CanonicalURL:    "https://example.com",
		H1:              []string{"Heading"},
		WordCount:       100,
	}

	addedImages := make([]types.Image, 6)
	for i := range addedImages {
		addedImages[i] = types.Image{Src: "https://example.com/img.png"}
	}
	diff := &types.CompareDiff{
		Images: &types.ImagesDiff{
			Added:   addedImages,
			Removed: []types.Image{},
		},
	}

	impact := ComputeRenderingImpact(jsData, nonJSData, diff)

	if impact.OverallChange != "major" {
		t.Errorf("expected overall_change='major', got %q", impact.OverallChange)
	}
	if impact.ImagesAdded != 6 {
		t.Errorf("expected ImagesAdded=6, got %d", impact.ImagesAdded)
	}
}

func TestComputeRenderingImpact_ContentChangePercent(t *testing.T) {
	tests := []struct {
		name     string
		jsWC     int
		nonJSWC  int
		expected float64
	}{
		{
			name:     "large difference",
			jsWC:     450,
			nonJSWC:  12,
			expected: 97.3,
		},
		{
			name:     "both zero",
			jsWC:     0,
			nonJSWC:  0,
			expected: 0.0,
		},
		{
			name:     "identical counts",
			jsWC:     100,
			nonJSWC:  100,
			expected: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsData := &types.RenderData{WordCount: tc.jsWC}
			nonJSData := &types.RenderData{WordCount: tc.nonJSWC}
			diff := &types.CompareDiff{}

			impact := ComputeRenderingImpact(jsData, nonJSData, diff)

			if impact.ContentChangePercent != tc.expected {
				t.Errorf("expected ContentChangePercent=%v, got %v", tc.expected, impact.ContentChangePercent)
			}
		})
	}
}

func TestComputeRenderingImpact_ZeroWordCounts(t *testing.T) {
	jsData := &types.RenderData{WordCount: 0}
	nonJSData := &types.RenderData{WordCount: 0}
	diff := &types.CompareDiff{}

	impact := ComputeRenderingImpact(jsData, nonJSData, diff)

	if impact.ContentChangePercent != 0.0 {
		t.Errorf("expected ContentChangePercent=0.0, got %f", impact.ContentChangePercent)
	}
}
