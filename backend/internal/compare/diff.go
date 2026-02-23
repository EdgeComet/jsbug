package compare

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/user/jsbug/internal/types"
)

// DiffString compares two string values and returns a StringDiff if they differ.
// Returns nil when both values are identical.
func DiffString(jsVal, nonJSVal string) *types.StringDiff {
	if jsVal == nonJSVal {
		return nil
	}
	return &types.StringDiff{JSValue: jsVal, NonJSValue: nonJSVal}
}

// DiffStringSlice performs a set-based comparison of two string slices.
// Added contains items present in js but not in nonJS.
// Removed contains items present in nonJS but not in js.
// Returns nil when there are no differences.
func DiffStringSlice(js, nonJS []string) *types.StringSliceDiff {
	jsSet := make(map[string]bool, len(js))
	for _, s := range js {
		jsSet[s] = true
	}
	nonJSSet := make(map[string]bool, len(nonJS))
	for _, s := range nonJS {
		nonJSSet[s] = true
	}

	added := []string{}
	for s := range jsSet {
		if !nonJSSet[s] {
			added = append(added, s)
		}
	}

	removed := []string{}
	for s := range nonJSSet {
		if !jsSet[s] {
			removed = append(removed, s)
		}
	}

	if len(added) == 0 && len(removed) == 0 {
		return nil
	}

	sort.Strings(added)
	sort.Strings(removed)

	return &types.StringSliceDiff{
		Added:   added,
		Removed: removed,
	}
}

// sectionKey builds a lookup key for a section based on heading level and text.
func sectionKey(level int, heading string) string {
	return fmt.Sprintf("h%d:%s", level, strings.ToLower(strings.TrimSpace(heading)))
}

// sectionTextKey returns just the normalized heading text for text-only matching.
func sectionTextKey(heading string) string {
	return strings.ToLower(strings.TrimSpace(heading))
}

// DiffSections compares two slices of sections and produces a list of SectionDiff entries.
// Sections are matched in two passes: first by exact key (heading level + text), then by
// text-only fallback for sections where JS rendering changed the heading level.
// Unmatched JS sections are marked "added_by_js"; unmatched non-JS sections are "removed_by_js".
func DiffSections(js, nonJS []types.Section) []types.SectionDiff {
	if len(js) == 0 && len(nonJS) == 0 {
		return nil
	}

	// Build maps of non-JS sections grouped by exact key and text key.
	nonJSByKey := make(map[string][]int, len(nonJS))
	nonJSByText := make(map[string][]int, len(nonJS))
	for i, s := range nonJS {
		key := sectionKey(s.HeadingLevel, s.HeadingText)
		nonJSByKey[key] = append(nonJSByKey[key], i)
		textKey := sectionTextKey(s.HeadingText)
		if textKey != "" {
			nonJSByText[textKey] = append(nonJSByText[textKey], i)
		}
	}

	consumed := make(map[int]bool, len(nonJS))
	jsMatched := make(map[int]bool, len(js))
	var result []types.SectionDiff

	// Pass 1: exact key match.
	for i, jsSection := range js {
		key := sectionKey(jsSection.HeadingLevel, jsSection.HeadingText)
		indices := nonJSByKey[key]
		matchIdx := -1
		for _, idx := range indices {
			if !consumed[idx] {
				matchIdx = idx
				break
			}
		}
		if matchIdx >= 0 {
			consumed[matchIdx] = true
			jsMatched[i] = true
			nonJSSection := nonJS[matchIdx]
			if jsSection.BodyMarkdown == nonJSSection.BodyMarkdown {
				result = append(result, types.SectionDiff{
					SectionID:    jsSection.SectionID,
					HeadingLevel: jsSection.HeadingLevel,
					HeadingText:  jsSection.HeadingText,
					Status:       "unchanged",
				})
			} else {
				result = append(result, types.SectionDiff{
					SectionID:         jsSection.SectionID,
					HeadingLevel:      jsSection.HeadingLevel,
					HeadingText:       jsSection.HeadingText,
					Status:            "changed",
					NonJSBodyMarkdown: nonJSSection.BodyMarkdown,
				})
			}
		}
	}

	// Pass 2: text-only fallback for unmatched JS sections.
	for i, jsSection := range js {
		if jsMatched[i] {
			continue
		}
		textKey := sectionTextKey(jsSection.HeadingText)
		if textKey == "" {
			continue // skip empty text (intro sections)
		}
		indices := nonJSByText[textKey]
		matchIdx := -1
		for _, idx := range indices {
			if !consumed[idx] {
				matchIdx = idx
				break
			}
		}
		if matchIdx >= 0 {
			consumed[matchIdx] = true
			jsMatched[i] = true
			nonJSSection := nonJS[matchIdx]
			result = append(result, types.SectionDiff{
				SectionID:           jsSection.SectionID,
				HeadingLevel:        jsSection.HeadingLevel,
				HeadingText:         jsSection.HeadingText,
				Status:              "changed",
				NonJSBodyMarkdown:   nonJSSection.BodyMarkdown,
				HeadingLevelChanged: true,
			})
		}
	}

	// Remaining unmatched JS sections -> added_by_js.
	for i, jsSection := range js {
		if !jsMatched[i] {
			result = append(result, types.SectionDiff{
				SectionID:    jsSection.SectionID,
				HeadingLevel: jsSection.HeadingLevel,
				HeadingText:  jsSection.HeadingText,
				Status:       "added_by_js",
			})
		}
	}

	// Remaining unconsumed non-JS sections -> removed_by_js.
	for i, s := range nonJS {
		if !consumed[i] {
			result = append(result, types.SectionDiff{
				SectionID:    s.SectionID,
				HeadingLevel: s.HeadingLevel,
				HeadingText:  s.HeadingText,
				Status:       "removed_by_js",
			})
		}
	}

	return result
}

// DiffLinks compares two slices of links, deduplicating by Href.
// Always returns a non-nil result with counts and sorted Added/Removed slices.
func DiffLinks(js, nonJS []types.Link) *types.LinksDiff {
	result := &types.LinksDiff{
		JSCount:    len(js),
		NonJSCount: len(nonJS),
		Added:      []types.Link{},
		Removed:    []types.Link{},
	}

	jsMap := make(map[string]types.Link, len(js))
	for _, l := range js {
		if _, exists := jsMap[l.Href]; !exists {
			jsMap[l.Href] = l
		}
	}

	nonJSMap := make(map[string]types.Link, len(nonJS))
	for _, l := range nonJS {
		if _, exists := nonJSMap[l.Href]; !exists {
			nonJSMap[l.Href] = l
		}
	}

	for href, l := range jsMap {
		if _, exists := nonJSMap[href]; !exists {
			result.Added = append(result.Added, l)
		}
	}

	for href, l := range nonJSMap {
		if _, exists := jsMap[href]; !exists {
			result.Removed = append(result.Removed, l)
		}
	}

	sort.Slice(result.Added, func(i, j int) bool {
		return result.Added[i].Href < result.Added[j].Href
	})
	sort.Slice(result.Removed, func(i, j int) bool {
		return result.Removed[i].Href < result.Removed[j].Href
	})

	return result
}

// DiffImages compares two slices of images, deduplicating by Src.
// Always returns a non-nil result with counts and sorted Added/Removed slices.
func DiffImages(js, nonJS []types.Image) *types.ImagesDiff {
	result := &types.ImagesDiff{
		JSCount:    len(js),
		NonJSCount: len(nonJS),
		Added:      []types.Image{},
		Removed:    []types.Image{},
	}

	jsMap := make(map[string]types.Image, len(js))
	for _, img := range js {
		if _, exists := jsMap[img.Src]; !exists {
			jsMap[img.Src] = img
		}
	}

	nonJSMap := make(map[string]types.Image, len(nonJS))
	for _, img := range nonJS {
		if _, exists := nonJSMap[img.Src]; !exists {
			nonJSMap[img.Src] = img
		}
	}

	for src, img := range jsMap {
		if _, exists := nonJSMap[src]; !exists {
			result.Added = append(result.Added, img)
		}
	}

	for src, img := range nonJSMap {
		if _, exists := jsMap[src]; !exists {
			result.Removed = append(result.Removed, img)
		}
	}

	sort.Slice(result.Added, func(i, j int) bool {
		return result.Added[i].Src < result.Added[j].Src
	})
	sort.Slice(result.Removed, func(i, j int) bool {
		return result.Removed[i].Src < result.Removed[j].Src
	})

	return result
}

// extractType extracts the @type from a JSON-LD block.
// Returns the @type as a string key, or a synthetic key for blocks without a valid @type.
func extractType(raw json.RawMessage, index int) string {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Sprintf("(unknown-%d)", index)
	}

	typeVal, ok := obj["@type"]
	if !ok {
		return fmt.Sprintf("(unknown-%d)", index)
	}

	switch v := typeVal.(type) {
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ",")
		}
		return fmt.Sprintf("(unknown-%d)", index)
	default:
		return fmt.Sprintf("(unknown-%d)", index)
	}
}

// uniqueKey returns a map key for a JSON-LD block, appending an index suffix
// if the base type key already exists in the map to avoid overwriting duplicates.
func uniqueKey(base string, m map[string]json.RawMessage) string {
	if _, exists := m[base]; !exists {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s#%d", base, i)
		if _, exists := m[candidate]; !exists {
			return candidate
		}
	}
}

// DiffStructuredData compares two slices of JSON-LD blocks by their @type field.
// Returns nil when there are no differences.
func DiffStructuredData(js, nonJS []json.RawMessage) *types.StructuredDataDiff {
	jsMap := make(map[string]json.RawMessage, len(js))
	for i, raw := range js {
		base := extractType(raw, i)
		key := uniqueKey(base, jsMap)
		jsMap[key] = raw
	}

	nonJSMap := make(map[string]json.RawMessage, len(nonJS))
	for i, raw := range nonJS {
		base := extractType(raw, i)
		key := uniqueKey(base, nonJSMap)
		nonJSMap[key] = raw
	}

	added := []string{}
	removed := []string{}
	changed := []string{}

	for key, jsRaw := range jsMap {
		nonJSRaw, exists := nonJSMap[key]
		if !exists {
			added = append(added, key)
		} else if string(jsRaw) != string(nonJSRaw) {
			changed = append(changed, key)
		}
	}

	for key := range nonJSMap {
		if _, exists := jsMap[key]; !exists {
			removed = append(removed, key)
		}
	}

	if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
		return nil
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)

	return &types.StructuredDataDiff{
		Added:   added,
		Removed: removed,
		Changed: changed,
	}
}
