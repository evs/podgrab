package db

import (
	"log"
	"os"
	"path"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

//DB is
var DB *gorm.DB

//Init is used to Initialize Database
func Init() (*gorm.DB, error) {
	configPath := os.Getenv("CONFIG")
	dbPath := path.Join(configPath, "podgrab.db")
	log.Println(dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	localDB, _ := db.DB()
	localDB.SetMaxIdleConns(10)
	// Enable WAL mode for better concurrent read performance
	localDB.Exec("PRAGMA journal_mode=WAL")
	localDB.Exec("PRAGMA busy_timeout=5000")
	DB = db
	return DB, nil
}

//Migrate Database
func Migrate() {
	DB.AutoMigrate(&Podcast{}, &PodcastItem{}, &Setting{}, &Migration{}, &JobLock{}, &Tag{})
	RunMigrations()
}

// Using this function to get a connection, you can create your connection pool here.
func GetDB() *gorm.DB {
	return DB
}
