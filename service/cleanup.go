package service

import (
	"time"

	"github.com/akhilrex/podgrab/db"
	"go.uber.org/zap"
)

// CleanupOldEpisodes removes episode files and DB records older than 90 days.
func CleanupOldEpisodes() error {
	const age = 90 * 24 * time.Hour
	cutoff := time.Now().Add(-age)

	items, err := db.GetOldPodcastItems(cutoff)
	if err != nil {
		Logger.Errorw("CleanupOldEpisodes: failed to fetch old items", zap.Error(err))
		return err
	}

	for _, item := range *items {
		if item.DownloadPath != "" {
			_ = DeleteFile(item.DownloadPath)
		}
		if item.LocalImage != "" {
			_ = DeleteFile(item.LocalImage)
		}
		_ = SetPodcastItemAsNotDownloaded(item.ID, db.Deleted)
	}

	Logger.Infow("CleanupOldEpisodes finished", "count", len(*items))
	return nil
}
