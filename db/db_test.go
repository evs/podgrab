package db

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	DB = db
	DB.AutoMigrate(&Podcast{}, &PodcastItem{}, &Setting{}, &Migration{}, &JobLock{}, &Tag{})
	RunMigrations()
	return db
}

func TestBasicDBSetup(t *testing.T) {
	setupTestDB(t)

	podcast := Podcast{
		Title: "Test Podcast",
		URL:   "https://example.com/feed.xml",
	}
	err := CreatePodcast(&podcast)
	if err != nil {
		t.Fatalf("CreatePodcast: %v", err)
	}
	if podcast.ID == "" {
		t.Errorf("podcast.ID is empty after create")
	}
	if _, err := uuid.Parse(podcast.ID); err != nil {
		t.Errorf("podcast.ID is not a valid UUID: %v", err)
	}
	if podcast.Title != "Test Podcast" {
		t.Errorf("podcast.Title = %q, want %q", podcast.Title, "Test Podcast")
	}

	var fetched Podcast
	err = GetPodcastById(podcast.ID, &fetched)
	if err != nil {
		t.Fatalf("GetPodcastById: %v", err)
	}
	if fetched.Title != "Test Podcast" {
		t.Errorf("fetched.Title = %q, want %q", fetched.Title, "Test Podcast")
	}

	err = DeletePodcastById(podcast.ID)
	if err != nil {
		t.Fatalf("DeletePodcastById: %v", err)
	}

	var afterDelete Podcast
	err = GetPodcastById(podcast.ID, &afterDelete)
	if err == nil {
		t.Errorf("podcast still exists after delete")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetPodcastByURL(t *testing.T) {
	setupTestDB(t)

	created := Podcast{
		Title: "URL Lookup Test",
		URL:   "https://example.com/findme.xml",
	}
	if err := CreatePodcast(&created); err != nil {
		t.Fatalf("CreatePodcast: %v", err)
	}

	var fetched Podcast
	if err := GetPodcastByURL(created.URL, &fetched); err != nil {
		t.Fatalf("GetPodcastByURL: %v", err)
	}
	if fetched.Title != created.Title {
		t.Errorf("fetched.Title = %q, want %q", fetched.Title, created.Title)
	}

	var notFound Podcast
	if err := GetPodcastByURL("https://example.com/nonexistent.xml", &notFound); err == nil {
		t.Errorf("GetPodcastByURL for nonexistent URL should return error, got nil")
	}
}

func TestGetAllPodcasts(t *testing.T) {
	setupTestDB(t)

	for i := 0; i < 3; i++ {
		p := Podcast{Title: "Podcast " + string(rune('A'+i)), URL: "https://example.com/podcast-" + string(rune('a'+i)) + ".xml"}
		if err := CreatePodcast(&p); err != nil {
			t.Fatalf("CreatePodcast %d: %v", i, err)
		}
	}

	var podcasts []Podcast
	if err := GetAllPodcasts(&podcasts, "created_at"); err != nil {
		t.Fatalf("GetAllPodcasts: %v", err)
	}
	// Tests share a :memory:?cache=shared DB; only verify our created podcasts are present
	found := make(map[string]bool)
	for _, p := range podcasts {
		found[p.Title] = true
	}
	for _, want := range []string{"Podcast A", "Podcast B", "Podcast C"} {
		if !found[want] {
			t.Fatalf("podcast %q not found in GetAllPodcasts results", want)
		}
	}
}

func TestPodcastItemsCRUD(t *testing.T) {
	setupTestDB(t)

	podcast := Podcast{
		Title: "Items Parent",
		URL:   "https://example.com/items-parent.xml",
	}
	if err := CreatePodcast(&podcast); err != nil {
		t.Fatalf("CreatePodcast: %v", err)
	}

	item1 := PodcastItem{
		PodcastID: podcast.ID,
		Title:     "Episode 1",
		GUID:      "guid-1",
		PubDate:   time.Now(),
	}
	item2 := PodcastItem{
		PodcastID: podcast.ID,
		Title:     "Episode 2",
		GUID:      "guid-2",
		PubDate:   time.Now(),
	}
	DB.Create(&item1)
	DB.Create(&item2)

	var items []PodcastItem
	if err := GetAllPodcastItemsByPodcastId(podcast.ID, &items); err != nil {
		t.Fatalf("GetAllPodcastItemsByPodcastId: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	if err := DeletePodcastItemById(item1.ID); err != nil {
		t.Fatalf("DeletePodcastItemById: %v", err)
	}

	var remaining []PodcastItem
	if err := GetAllPodcastItemsByPodcastId(podcast.ID, &remaining); err != nil {
		t.Fatalf("GetAllPodcastItemsByPodcastId (after delete): %v", err)
	}
	if len(remaining) != 1 {
		t.Errorf("len(remaining) = %d, want 1 after deleting one item", len(remaining))
	}
}
