package service

import (
	"testing"

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
