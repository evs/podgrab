package service

import (
	"testing"
	"time"

	"github.com/akhilrex/podgrab/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.DB = gdb
	db.DB.AutoMigrate(&db.Podcast{}, &db.PodcastItem{}, &db.Setting{}, &db.Migration{}, &db.JobLock{}, &db.Tag{})
	db.RunMigrations()
	return gdb
}

func TestParseOpml(t *testing.T) {
	opml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head>
    <title>Test Subscriptions</title>
  </head>
  <body>
    <outline text="A Podcast" type="rss" xmlUrl="https://example.com/feed.xml"/>
  </body>
</opml>`
	result, err := ParseOpml(opml)
	if err != nil {
		t.Fatalf("ParseOpml: %v", err)
	}
	if result.Head.Title != "Test Subscriptions" {
		t.Errorf("Head.Title = %q, want %q", result.Head.Title, "Test Subscriptions")
	}
	if len(result.Body.Outline) != 1 {
		t.Fatalf("len(Outline) = %d, want 1", len(result.Body.Outline))
	}
	if result.Body.Outline[0].AttrText != "A Podcast" {
		t.Errorf("Outline[0].AttrText = %q, want %q", result.Body.Outline[0].AttrText, "A Podcast")
	}
}

func TestServiceTestDBSetup(t *testing.T) {
	setupServiceTestDB(t)

	podcast := db.Podcast{
		Title: "Service Test Podcast",
		URL:   "https://example.com/service-feed.xml",
	}
	err := db.CreatePodcast(&podcast)
	if err != nil {
		t.Fatalf("CreatePodcast: %v", err)
	}
	if podcast.ID == "" {
		t.Errorf("podcast.ID is empty after create")
	}

	fetched := GetPodcastById(podcast.ID)
	if fetched == nil {
		t.Fatal("GetPodcastById returned nil")
	}
	if fetched.Title != "Service Test Podcast" {
		t.Errorf("fetched.Title = %q, want %q", fetched.Title, "Service Test Podcast")
	}

	err = db.DeletePodcastById(podcast.ID)
	if err != nil {
		t.Fatalf("DeletePodcastById: %v", err)
	}
}

func TestGetPodcastById_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	result := GetPodcastById("00000000-0000-0000-0000-000000000000")
	if result == nil {
		t.Error("GetPodcastById returned nil for nonexistent ID (expected non-nil zero-value)")
	}
	if result.Title != "" {
		t.Errorf("GetPodcastById for nonexistent ID has Title = %q, want empty string", result.Title)
	}
}

func TestParseRSSDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		year    int
	}{
		{"RFC1123", "Mon, 02 Jan 2006 15:04:05 MST", false, 2006},
		{"RFC1123Z", "Mon, 02 Jan 2006 15:04:05 -0700", false, 2006},
		{"RFC3339", "2006-01-02T15:04:05Z", false, 2006},
		{"RFC3339Nano", "2006-01-02T15:04:05.999999999Z", false, 2006},
		{"ISO8601", "2006-01-02T15:04:05+00:00", false, 2006},
		{"ModifiedRFC1123", "Mon, 2 Jan 2006 15:04:05 MST", false, 2006},
		{"ModifiedRFC1123Z", "Mon, 2 Jan 2006 15:04:05 -0700", false, 2006},
		{"empty", "", true, 0},
		{"garbage", "not-a-date", true, 0},
		{"RFC822", "02 Jan 06 15:04 MST", false, 2006},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRSSDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRSSDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got.Year() != tt.year {
				t.Errorf("parseRSSDate(%q) year = %d, want %d", tt.input, got.Year(), tt.year)
			}
		})
	}
}

func TestGetAllPodcasts_Service(t *testing.T) {
	setupServiceTestDB(t)

	podcasts := []db.Podcast{
		{Title: "Svc Podcast X", URL: "https://svc-x.example.com/feed.xml"},
		{Title: "Svc Podcast Y", URL: "https://svc-y.example.com/feed.xml"},
	}
	for i := range podcasts {
		err := db.CreatePodcast(&podcasts[i])
		if err != nil {
			t.Fatalf("CreatePodcast %q: %v", podcasts[i].Title, err)
		}
	}

	results := GetAllPodcasts("created_at")
	if results == nil {
		t.Fatal("GetAllPodcasts returned nil")
	}
	if len(*results) < 2 {
		t.Fatalf("len(results) = %d, want >= 2", len(*results))
	}
	found := make(map[string]bool)
	for _, p := range *results {
		found[p.Title] = true
	}
	for _, want := range []string{"Svc Podcast X", "Svc Podcast Y"} {
		if !found[want] {
			t.Errorf("podcast %q not found in service layer results", want)
		}
	}
}

func TestDeletePodcastEpisodes_ClearsItems(t *testing.T) {
	setupServiceTestDB(t)

	podcast := db.Podcast{
		Title: "Delete Episodes Test",
		URL:   "https://example.com/delete-episodes.xml",
	}
	err := db.CreatePodcast(&podcast)
	if err != nil {
		t.Fatalf("CreatePodcast: %v", err)
	}

	item1 := db.PodcastItem{
		PodcastID:      podcast.ID,
		Title:          "Ep 1",
		GUID:           "delete-ep-guid-1",
		PubDate:        time.Now(),
		DownloadStatus: db.Downloaded,
		DownloadPath:   "/nonexistent/path/ep1.mp3",
	}
	item2 := db.PodcastItem{
		PodcastID:      podcast.ID,
		Title:          "Ep 2",
		GUID:           "delete-ep-guid-2",
		PubDate:        time.Now(),
		DownloadStatus: db.Downloaded,
		DownloadPath:   "/nonexistent/path/ep2.mp3",
	}
	db.DB.Create(&item1)
	db.DB.Create(&item2)

	err = DeletePodcastEpisodes(podcast.ID)
	// DeletePodcastEpisodes may error trying to DeleteFile on nonexistent paths
	_ = err

	var remainingItems []db.PodcastItem
	err = db.GetAllPodcastItemsByPodcastId(podcast.ID, &remainingItems)
	if err != nil {
		t.Fatalf("GetAllPodcastItemsByPodcastId: %v", err)
	}

	for _, item := range remainingItems {
		if item.DownloadStatus != db.Deleted {
			t.Errorf("item %q (GUID=%q) has DownloadStatus=%d, want Deleted(%d)",
				item.Title, item.GUID, item.DownloadStatus, db.Deleted)
		}
	}
}
