package compare

import (
	"math"
	"sort"

	"github.com/user/jsbug/internal/types"
)

// Threshold constants for classifying overall rendering impact.
const (
	MinorContentThreshold float64 = 5.0
	MajorContentThreshold float64 = 30.0
	MajorLinksThreshold   int     = 10
	MajorImagesThreshold  int     = 5
)

// ComputeRenderingImpact computes the rendering impact summary by comparing
// JS-rendered data against non-JS data and the computed diff.
func ComputeRenderingImpact(jsData, nonJSData *types.RenderData, diff *types.CompareDiff) *types.RenderingImpact {
	impact := &types.RenderingImpact{
		TitleChanged:     jsData.Title != nonJSData.Title,
		MetaDescChanged:  jsData.MetaDescription != nonJSData.MetaDescription,
		CanonicalChanged: jsData.CanonicalURL != nonJSData.CanonicalURL,
		H1Changed:        !slicesEqual(jsData.H1, nonJSData.H1),
		WordCountJS:      jsData.WordCount,
		WordCountNonJS:   nonJSData.WordCount,
	}

	// Content change percent
	impact.ContentChangePercent = contentChangePercent(jsData.WordCount, nonJSData.WordCount)

	// Links
	if diff.Links != nil {
		impact.LinksAdded = len(diff.Links.Added)
		impact.LinksRemoved = len(diff.Links.Removed)
	}

	// Images
	if diff.Images != nil {
		impact.ImagesAdded = len(diff.Images.Added)
		impact.ImagesRemoved = len(diff.Images.Removed)
	}

	// Structured data
	if diff.StructuredData != nil {
		impact.StructuredDataAdded = len(diff.StructuredData.Added)
		impact.StructuredDataRemoved = len(diff.StructuredData.Removed)
	}

	impact.OverallChange = classifyOverallChange(impact)

	return impact
}

// classifyOverallChange determines whether the rendering impact is "none", "minor", or "major".
func classifyOverallChange(impact *types.RenderingImpact) string {
	if !impact.TitleChanged &&
		!impact.MetaDescChanged &&
		!impact.CanonicalChanged &&
		!impact.H1Changed &&
		impact.ContentChangePercent < MinorContentThreshold &&
		impact.LinksAdded == 0 &&
		impact.LinksRemoved == 0 &&
		impact.ImagesAdded == 0 &&
		impact.ImagesRemoved == 0 &&
		impact.StructuredDataAdded == 0 &&
		impact.StructuredDataRemoved == 0 {
		return "none"
	}

	if impact.TitleChanged ||
		impact.ContentChangePercent > MajorContentThreshold ||
		impact.LinksAdded > MajorLinksThreshold ||
		impact.ImagesAdded > MajorImagesThreshold {
		return "major"
	}

	return "minor"
}

// contentChangePercent computes the percentage change in word count between
// JS and non-JS versions, rounded to 1 decimal place.
// Formula: abs(js - nonJS) / max(js, nonJS, 1) * 100
func contentChangePercent(wordCountJS, wordCountNonJS int) float64 {
	diff := math.Abs(float64(wordCountJS) - float64(wordCountNonJS))
	maxCount := max(wordCountJS, wordCountNonJS, 1)
	result := diff / float64(maxCount) * 100
	return math.Round(result*10) / 10
}

// slicesEqual compares two string slices for equality after sorting copies.
// Nil and empty slices are considered equal.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}

	aCopy := make([]string, len(a))
	copy(aCopy, a)
	sort.Strings(aCopy)

	bCopy := make([]string, len(b))
	copy(bCopy, b)
	sort.Strings(bCopy)

	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	return true
}
