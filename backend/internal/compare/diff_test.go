package compare

import (
	"encoding/json"
	"testing"

	"github.com/user/jsbug/internal/types"
)

func TestDiffString(t *testing.T) {
	tests := []struct {
		name     string
		jsVal    string
		nonJSVal string
		wantNil  bool
	}{
		{
			name:     "both_empty",
			jsVal:    "",
			nonJSVal: "",
			wantNil:  true,
		},
		{
			name:     "both_same",
			jsVal:    "Hello World",
			nonJSVal: "Hello World",
			wantNil:  true,
		},
		{
			name:     "different",
			jsVal:    "JS Title",
			nonJSVal: "Non-JS Title",
			wantNil:  false,
		},
		{
			name:     "js_empty",
			jsVal:    "",
			nonJSVal: "Non-JS Title",
			wantNil:  false,
		},
		{
			name:     "non_js_empty",
			jsVal:    "JS Title",
			nonJSVal: "",
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DiffString(tt.jsVal, tt.nonJSVal)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
			} else {
				if got == nil {
					t.Fatal("expected non-nil result")
				}
				if got.JSValue != tt.jsVal {
					t.Errorf("JSValue = %q, want %q", got.JSValue, tt.jsVal)
				}
				if got.NonJSValue != tt.nonJSVal {
					t.Errorf("NonJSValue = %q, want %q", got.NonJSValue, tt.nonJSVal)
				}
			}
		})
	}
}

func TestDiffStringSlice(t *testing.T) {
	tests := []struct {
		name        string
		js          []string
		nonJS       []string
		wantNil     bool
		wantAdded   []string
		wantRemoved []string
	}{
		{
			name:    "both_nil",
			js:      nil,
			nonJS:   nil,
			wantNil: true,
		},
		{
			name:    "both_same",
			js:      []string{"a", "b"},
			nonJS:   []string{"a", "b"},
			wantNil: true,
		},
		{
			name:        "added_only",
			js:          []string{"a", "b", "c"},
			nonJS:       []string{"a"},
			wantNil:     false,
			wantAdded:   []string{"b", "c"},
			wantRemoved: []string{},
		},
		{
			name:        "removed_only",
			js:          []string{"a"},
			nonJS:       []string{"a", "b", "c"},
			wantNil:     false,
			wantAdded:   []string{},
			wantRemoved: []string{"b", "c"},
		},
		{
			name:        "both_added_and_removed",
			js:          []string{"a", "c"},
			nonJS:       []string{"a", "b"},
			wantNil:     false,
			wantAdded:   []string{"c"},
			wantRemoved: []string{"b"},
		},
		{
			name:        "sorted_output",
			js:          []string{"zebra", "apple", "mango"},
			nonJS:       []string{"banana", "cherry"},
			wantNil:     false,
			wantAdded:   []string{"apple", "mango", "zebra"},
			wantRemoved: []string{"banana", "cherry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DiffStringSlice(tt.js, tt.nonJS)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if !sliceEqual(got.Added, tt.wantAdded) {
				t.Errorf("Added = %v, want %v", got.Added, tt.wantAdded)
			}
			if !sliceEqual(got.Removed, tt.wantRemoved) {
				t.Errorf("Removed = %v, want %v", got.Removed, tt.wantRemoved)
			}
		})
	}
}

func TestDiffSections_Unchanged(t *testing.T) {
	sections := []types.Section{
		{SectionID: "s1", HeadingLevel: 1, HeadingText: "Intro", BodyMarkdown: "Hello world"},
		{SectionID: "s2", HeadingLevel: 2, HeadingText: "Details", BodyMarkdown: "Some details"},
	}

	result := DiffSections(sections, sections)
	if len(result) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(result))
	}
	for _, d := range result {
		if d.Status != "unchanged" {
			t.Errorf("section %q: expected status unchanged, got %q", d.SectionID, d.Status)
		}
		if d.NonJSBodyMarkdown != "" {
			t.Errorf("section %q: expected empty NonJSBodyMarkdown, got %q", d.SectionID, d.NonJSBodyMarkdown)
		}
	}
}

func TestDiffSections_AddedByJS(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 1, HeadingText: "Intro", BodyMarkdown: "Hello"},
		{SectionID: "s2", HeadingLevel: 2, HeadingText: "Extra", BodyMarkdown: "JS only"},
	}
	nonJS := []types.Section{
		{SectionID: "s1", HeadingLevel: 1, HeadingText: "Intro", BodyMarkdown: "Hello"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(result))
	}
	if result[0].Status != "unchanged" {
		t.Errorf("first section: expected unchanged, got %q", result[0].Status)
	}
	if result[1].Status != "added_by_js" {
		t.Errorf("second section: expected added_by_js, got %q", result[1].Status)
	}
	if result[1].SectionID != "s2" {
		t.Errorf("second section: expected SectionID s2, got %q", result[1].SectionID)
	}
}

func TestDiffSections_RemovedByJS(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 1, HeadingText: "Intro", BodyMarkdown: "Hello"},
	}
	nonJS := []types.Section{
		{SectionID: "s1", HeadingLevel: 1, HeadingText: "Intro", BodyMarkdown: "Hello"},
		{SectionID: "s2", HeadingLevel: 2, HeadingText: "Old Section", BodyMarkdown: "Old content"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(result))
	}
	if result[0].Status != "unchanged" {
		t.Errorf("first section: expected unchanged, got %q", result[0].Status)
	}
	if result[1].Status != "removed_by_js" {
		t.Errorf("second section: expected removed_by_js, got %q", result[1].Status)
	}
	if result[1].SectionID != "s2" {
		t.Errorf("second section: expected SectionID s2, got %q", result[1].SectionID)
	}
}

func TestDiffSections_Changed(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 2, HeadingText: "About", BodyMarkdown: "Updated content"},
	}
	nonJS := []types.Section{
		{SectionID: "s1-old", HeadingLevel: 2, HeadingText: "About", BodyMarkdown: "Original content"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result))
	}
	if result[0].Status != "changed" {
		t.Errorf("expected status changed, got %q", result[0].Status)
	}
	if result[0].SectionID != "s1" {
		t.Errorf("expected SectionID from JS section (s1), got %q", result[0].SectionID)
	}
	if result[0].NonJSBodyMarkdown != "Original content" {
		t.Errorf("expected NonJSBodyMarkdown = %q, got %q", "Original content", result[0].NonJSBodyMarkdown)
	}
}

func TestDiffSections_DuplicateHeadings(t *testing.T) {
	js := []types.Section{
		{SectionID: "js1", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "First JS"},
		{SectionID: "js2", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "Second JS"},
	}
	nonJS := []types.Section{
		{SectionID: "njs1", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "First NonJS"},
		{SectionID: "njs2", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "Second NonJS"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(result))
	}

	// First JS section matches first nonJS section (first-match-wins).
	if result[0].SectionID != "js1" {
		t.Errorf("first diff: expected SectionID js1, got %q", result[0].SectionID)
	}
	if result[0].Status != "changed" {
		t.Errorf("first diff: expected changed, got %q", result[0].Status)
	}
	if result[0].NonJSBodyMarkdown != "First NonJS" {
		t.Errorf("first diff: expected NonJSBodyMarkdown %q, got %q", "First NonJS", result[0].NonJSBodyMarkdown)
	}

	// Second JS section matches second nonJS section.
	if result[1].SectionID != "js2" {
		t.Errorf("second diff: expected SectionID js2, got %q", result[1].SectionID)
	}
	if result[1].Status != "changed" {
		t.Errorf("second diff: expected changed, got %q", result[1].Status)
	}
	if result[1].NonJSBodyMarkdown != "Second NonJS" {
		t.Errorf("second diff: expected NonJSBodyMarkdown %q, got %q", "Second NonJS", result[1].NonJSBodyMarkdown)
	}
}

func TestDiffSections_IntroSection(t *testing.T) {
	js := []types.Section{
		{SectionID: "intro-js", HeadingLevel: 0, HeadingText: "", BodyMarkdown: "JS intro"},
	}
	nonJS := []types.Section{
		{SectionID: "intro-nojs", HeadingLevel: 0, HeadingText: "", BodyMarkdown: "Non-JS intro"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result))
	}
	if result[0].Status != "changed" {
		t.Errorf("expected status changed, got %q", result[0].Status)
	}
	if result[0].SectionID != "intro-js" {
		t.Errorf("expected SectionID from JS (intro-js), got %q", result[0].SectionID)
	}
	if result[0].NonJSBodyMarkdown != "Non-JS intro" {
		t.Errorf("expected NonJSBodyMarkdown %q, got %q", "Non-JS intro", result[0].NonJSBodyMarkdown)
	}
}

func TestDiffSections_Empty(t *testing.T) {
	result := DiffSections(nil, nil)
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}

	result = DiffSections([]types.Section{}, []types.Section{})
	if result != nil {
		t.Errorf("expected nil for empty slices, got %+v", result)
	}
}

func TestDiffLinks_NoChanges(t *testing.T) {
	links := []types.Link{
		{Href: "https://example.com/a", Text: "A"},
		{Href: "https://example.com/b", Text: "B"},
	}

	result := DiffLinks(links, links)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.JSCount != 2 {
		t.Errorf("JSCount = %d, want 2", result.JSCount)
	}
	if result.NonJSCount != 2 {
		t.Errorf("NonJSCount = %d, want 2", result.NonJSCount)
	}
	if len(result.Added) != 0 {
		t.Errorf("expected empty Added, got %d items", len(result.Added))
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected empty Removed, got %d items", len(result.Removed))
	}
}

func TestDiffLinks_AddedAndRemoved(t *testing.T) {
	js := []types.Link{
		{Href: "https://example.com/a", Text: "A"},
		{Href: "https://example.com/c", Text: "C"},
		{Href: "https://example.com/d", Text: "D"},
	}
	nonJS := []types.Link{
		{Href: "https://example.com/a", Text: "A"},
		{Href: "https://example.com/b", Text: "B"},
	}

	result := DiffLinks(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.JSCount != 3 {
		t.Errorf("JSCount = %d, want 3", result.JSCount)
	}
	if result.NonJSCount != 2 {
		t.Errorf("NonJSCount = %d, want 2", result.NonJSCount)
	}
	if len(result.Added) != 2 {
		t.Fatalf("expected 2 added, got %d", len(result.Added))
	}
	// Verify alphabetical sort.
	if result.Added[0].Href != "https://example.com/c" {
		t.Errorf("Added[0].Href = %q, want https://example.com/c", result.Added[0].Href)
	}
	if result.Added[1].Href != "https://example.com/d" {
		t.Errorf("Added[1].Href = %q, want https://example.com/d", result.Added[1].Href)
	}
	if len(result.Removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(result.Removed))
	}
	if result.Removed[0].Href != "https://example.com/b" {
		t.Errorf("Removed[0].Href = %q, want https://example.com/b", result.Removed[0].Href)
	}
}

func TestDiffLinks_DuplicateHrefs(t *testing.T) {
	js := []types.Link{
		{Href: "https://example.com/a", Text: "First A"},
		{Href: "https://example.com/a", Text: "Second A"},
		{Href: "https://example.com/b", Text: "B"},
	}
	nonJS := []types.Link{
		{Href: "https://example.com/a", Text: "A"},
	}

	result := DiffLinks(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.JSCount != 3 {
		t.Errorf("JSCount = %d, want 3 (raw count)", result.JSCount)
	}
	if len(result.Added) != 1 {
		t.Fatalf("expected 1 added (unique), got %d", len(result.Added))
	}
	if result.Added[0].Href != "https://example.com/b" {
		t.Errorf("Added[0].Href = %q, want https://example.com/b", result.Added[0].Href)
	}
	// First occurrence wins for JS dedup.
	// The shared href "a" should not appear in Added since it exists in both.
	if len(result.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(result.Removed))
	}
}

func TestDiffImages_AddedAndRemoved(t *testing.T) {
	js := []types.Image{
		{Src: "https://example.com/img1.png", Alt: "Image 1"},
		{Src: "https://example.com/img3.png", Alt: "Image 3"},
	}
	nonJS := []types.Image{
		{Src: "https://example.com/img1.png", Alt: "Image 1"},
		{Src: "https://example.com/img2.png", Alt: "Image 2"},
	}

	result := DiffImages(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.JSCount != 2 {
		t.Errorf("JSCount = %d, want 2", result.JSCount)
	}
	if result.NonJSCount != 2 {
		t.Errorf("NonJSCount = %d, want 2", result.NonJSCount)
	}
	if len(result.Added) != 1 {
		t.Fatalf("expected 1 added, got %d", len(result.Added))
	}
	if result.Added[0].Src != "https://example.com/img3.png" {
		t.Errorf("Added[0].Src = %q, want https://example.com/img3.png", result.Added[0].Src)
	}
	if len(result.Removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(result.Removed))
	}
	if result.Removed[0].Src != "https://example.com/img2.png" {
		t.Errorf("Removed[0].Src = %q, want https://example.com/img2.png", result.Removed[0].Src)
	}
}

func TestDiffStructuredData_Added(t *testing.T) {
	js := []json.RawMessage{
		json.RawMessage(`{"@type":"Product","name":"Widget"}`),
	}
	var nonJS []json.RawMessage

	result := DiffStructuredData(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Added) != 1 || result.Added[0] != "Product" {
		t.Errorf("Added = %v, want [Product]", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", result.Removed)
	}
	if len(result.Changed) != 0 {
		t.Errorf("Changed = %v, want empty", result.Changed)
	}
}

func TestDiffStructuredData_Removed(t *testing.T) {
	var js []json.RawMessage
	nonJS := []json.RawMessage{
		json.RawMessage(`{"@type":"Organization","name":"Acme"}`),
	}

	result := DiffStructuredData(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Removed) != 1 || result.Removed[0] != "Organization" {
		t.Errorf("Removed = %v, want [Organization]", result.Removed)
	}
	if len(result.Added) != 0 {
		t.Errorf("Added = %v, want empty", result.Added)
	}
	if len(result.Changed) != 0 {
		t.Errorf("Changed = %v, want empty", result.Changed)
	}
}

func TestDiffStructuredData_Changed(t *testing.T) {
	js := []json.RawMessage{
		json.RawMessage(`{"@type":"Organization","name":"New Acme"}`),
	}
	nonJS := []json.RawMessage{
		json.RawMessage(`{"@type":"Organization","name":"Old Acme"}`),
	}

	result := DiffStructuredData(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Changed) != 1 || result.Changed[0] != "Organization" {
		t.Errorf("Changed = %v, want [Organization]", result.Changed)
	}
	if len(result.Added) != 0 {
		t.Errorf("Added = %v, want empty", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", result.Removed)
	}
}

func TestDiffStructuredData_NoType(t *testing.T) {
	js := []json.RawMessage{
		json.RawMessage(`{"name":"Widget"}`),
	}
	var nonJS []json.RawMessage

	result := DiffStructuredData(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Added) != 1 || result.Added[0] != "(unknown-0)" {
		t.Errorf("Added = %v, want [(unknown-0)]", result.Added)
	}
}

func TestDiffStructuredData_ArrayType(t *testing.T) {
	js := []json.RawMessage{
		json.RawMessage(`{"@type":["Product","ItemPage"],"name":"Widget"}`),
	}
	var nonJS []json.RawMessage

	result := DiffStructuredData(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Added) != 1 || result.Added[0] != "Product,ItemPage" {
		t.Errorf("Added = %v, want [Product,ItemPage]", result.Added)
	}
}

func TestDiffStructuredData_DuplicateType(t *testing.T) {
	js := []json.RawMessage{
		json.RawMessage(`{"@type":"Product","name":"Widget A"}`),
		json.RawMessage(`{"@type":"Product","name":"Widget B"}`),
	}
	nonJS := []json.RawMessage{
		json.RawMessage(`{"@type":"Product","name":"Widget A"}`),
	}

	result := DiffStructuredData(js, nonJS)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Added) != 1 {
		t.Errorf("Added = %v, want 1 entry for second Product", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", result.Removed)
	}
}

func TestDiffStructuredData_DuplicateTypeBothSides(t *testing.T) {
	js := []json.RawMessage{
		json.RawMessage(`{"@type":"Product","name":"Widget A"}`),
		json.RawMessage(`{"@type":"Product","name":"Widget B"}`),
	}
	nonJS := []json.RawMessage{
		json.RawMessage(`{"@type":"Product","name":"Widget A"}`),
		json.RawMessage(`{"@type":"Product","name":"Widget B"}`),
	}

	result := DiffStructuredData(js, nonJS)
	if result != nil {
		t.Errorf("expected nil (no differences), got %+v", result)
	}
}

func TestDiffSections_CrossLevelMatch(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "JS content"},
	}
	nonJS := []types.Section{
		{SectionID: "s1", HeadingLevel: 3, HeadingText: "Features", BodyMarkdown: "HTTP content"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result))
	}
	if result[0].Status != "changed" {
		t.Errorf("expected status changed, got %q", result[0].Status)
	}
	if !result[0].HeadingLevelChanged {
		t.Error("expected HeadingLevelChanged=true")
	}
	if result[0].HeadingLevel != 2 {
		t.Errorf("expected HeadingLevel=2 (from JS), got %d", result[0].HeadingLevel)
	}
	if result[0].NonJSBodyMarkdown != "HTTP content" {
		t.Errorf("expected NonJSBodyMarkdown=%q, got %q", "HTTP content", result[0].NonJSBodyMarkdown)
	}
}

func TestDiffSections_CrossLevelMatchSameBody(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "Same content"},
	}
	nonJS := []types.Section{
		{SectionID: "s1", HeadingLevel: 3, HeadingText: "Features", BodyMarkdown: "Same content"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result))
	}
	if result[0].Status != "changed" {
		t.Errorf("expected status changed, got %q", result[0].Status)
	}
	if !result[0].HeadingLevelChanged {
		t.Error("expected HeadingLevelChanged=true even though body is identical")
	}
}

func TestDiffSections_ExactMatchTakesPriority(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "A"},
		{SectionID: "s2", HeadingLevel: 3, HeadingText: "Features", BodyMarkdown: "B"},
	}
	nonJS := []types.Section{
		{SectionID: "s1", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "C"},
		{SectionID: "s2", HeadingLevel: 3, HeadingText: "Features", BodyMarkdown: "D"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(result))
	}
	for i, d := range result {
		if d.Status != "changed" {
			t.Errorf("result[%d]: expected status changed, got %q", i, d.Status)
		}
		if d.HeadingLevelChanged {
			t.Errorf("result[%d]: expected HeadingLevelChanged=false (exact key match)", i)
		}
	}
}

func TestDiffSections_TextFallbackSkipsEmptyText(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 0, HeadingText: "", BodyMarkdown: "JS intro"},
		{SectionID: "s2", HeadingLevel: 2, HeadingText: "Title", BodyMarkdown: "JS body"},
	}
	nonJS := []types.Section{
		{SectionID: "s1", HeadingLevel: 0, HeadingText: "", BodyMarkdown: "HTTP intro"},
		{SectionID: "s2", HeadingLevel: 3, HeadingText: "Title", BodyMarkdown: "HTTP body"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(result))
	}

	// Intro sections match by exact key in pass 1.
	introFound := false
	titleFound := false
	for _, d := range result {
		if d.SectionID == "s1" {
			introFound = true
			if d.Status != "changed" {
				t.Errorf("intro section: expected status changed, got %q", d.Status)
			}
			if d.HeadingLevelChanged {
				t.Error("intro section: expected HeadingLevelChanged=false (exact key match)")
			}
		}
		if d.SectionID == "s2" {
			titleFound = true
			if d.Status != "changed" {
				t.Errorf("title section: expected status changed, got %q", d.Status)
			}
			if !d.HeadingLevelChanged {
				t.Error("title section: expected HeadingLevelChanged=true (text fallback match)")
			}
		}
	}
	if !introFound {
		t.Error("intro section (s1) not found in results")
	}
	if !titleFound {
		t.Error("title section (s2) not found in results")
	}
}

func TestDiffSections_TextFallbackMultipleSameText(t *testing.T) {
	js := []types.Section{
		{SectionID: "s1", HeadingLevel: 2, HeadingText: "Features", BodyMarkdown: "A"},
	}
	nonJS := []types.Section{
		{SectionID: "s1", HeadingLevel: 3, HeadingText: "Features", BodyMarkdown: "C"},
		{SectionID: "s2", HeadingLevel: 4, HeadingText: "Features", BodyMarkdown: "D"},
	}

	result := DiffSections(js, nonJS)
	if len(result) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(result))
	}

	// First result: JS s1 matched NonJS s1 by text fallback.
	if result[0].SectionID != "s1" {
		t.Errorf("result[0]: expected SectionID s1, got %q", result[0].SectionID)
	}
	if result[0].Status != "changed" {
		t.Errorf("result[0]: expected status changed, got %q", result[0].Status)
	}
	if !result[0].HeadingLevelChanged {
		t.Error("result[0]: expected HeadingLevelChanged=true")
	}

	// Second result: NonJS s2 is unconsumed -> removed_by_js.
	if result[1].SectionID != "s2" {
		t.Errorf("result[1]: expected SectionID s2, got %q", result[1].SectionID)
	}
	if result[1].Status != "removed_by_js" {
		t.Errorf("result[1]: expected status removed_by_js, got %q", result[1].Status)
	}
}

// sliceEqual compares two string slices for equality.
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
