package db

import (
	"os"
	"testing"

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
