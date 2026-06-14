package service

import "testing"

// This test passing from the service package's working directory is the point:
// the embedded feed loads regardless of CWD (the old os.Getwd path could not).
func TestNewsService_LoadsEmbeddedFeed(t *testing.T) {
	items, err := NewNewsService().GetNews()
	if err != nil {
		t.Fatalf("GetNews: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one news item from the embedded feed")
	}
	if len(items) > maxNewsItems {
		t.Errorf("returned %d items, want at most %d", len(items), maxNewsItems)
	}
	if items[0].Title == "" {
		t.Error("first news item has no title")
	}
}
