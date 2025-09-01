package main

import "testing"

func TestRemoveEmptyElementsFromMetadataList(t *testing.T) {
	testCases := []struct {
		input        string
		expectedTags []any
	}{
		{
			input: `---
tags:
- golang
- 
-  web
- backend 
- 
---
Content here`,
			expectedTags: []any{"golang", "web", "backend"},
		},
		{
			input: `---
tags: 
- 
---
Content here`,
			expectedTags: []any{},
		},
	}

	for _, tc := range testCases {
		metadata, content, err := ParseMetadataAndContent([]byte(tc.input))
		if err != nil {
			t.Fatalf("Error parsing metadata: %v", err)
		}
		t.Log("Parsed metadata:", metadata)

		if content != "Content here" {
			t.Errorf("Expected content 'Content here', got '%s'", content)
		}

		tags, ok := metadata["tags"].([]any)
		if !ok {
			t.Fatalf("Expected tags to be a list, got %T", metadata["tags"])
		}

		if len(tags) != len(tc.expectedTags) {
			t.Errorf("Expected %d tags, got %d", len(tc.expectedTags), len(tags))
			continue
		}

		for i, tag := range tags {
			if tag != tc.expectedTags[i] {
				t.Errorf("Expected tag '%v', got '%v'", tc.expectedTags[i], tag)
			}
		}
	}
}
