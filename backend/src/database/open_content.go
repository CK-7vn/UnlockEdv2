package database

import (
	"UnlockEdv2/src/models"

	log "github.com/sirupsen/logrus"
)

func (db *DB) GetOpenContent(all bool) ([]models.OpenContentProvider, error) {
	var content []models.OpenContentProvider
	tx := db.Model(&models.OpenContentProvider{})
	if !all {
		tx.Where("currently_enabled = true")
	}
	if err := tx.Find(&content).Error; err != nil {
		return nil, newGetRecordsDBError(err, "open_content_providers")
	}
	return content, nil
}

func (db *DB) ToggleContentProvider(id int) error {
	var provider models.OpenContentProvider
	if err := db.Find(&provider, "id = ?", id).Error; err != nil {
		log.Errorln("unable to find conent provider with that ID")
		return newNotFoundDBError(err, "open_content_providers")
	}
	provider.CurrentlyEnabled = !provider.CurrentlyEnabled
	if err := db.Save(&provider).Error; err != nil {
		return newUpdateDBError(err, "open_content_providers")
	}
	return nil
}

func (db *DB) UpdateOpenContentProvider(prov *models.OpenContentProvider) error {
	if err := db.Save(prov).Error; err != nil {
		return newUpdateDBError(err, "open_content_providers")
	}
	return nil
}

func (db *DB) CreateContentProvider(provider *models.OpenContentProvider) error {
	if err := db.Create(&provider).Error; err != nil {
		return newCreateDBError(err, "open_content_providers")
	}
	return nil
}

func (db *DB) FindKolibriInstance() (*models.ProviderPlatform, error) {
	kolibri := models.ProviderPlatform{}
	if err := db.First(&kolibri, "type = ?", "kolibri").Error; err != nil {
		log.Error("error getting kolibri provider platform")
		return nil, newNotFoundDBError(err, "provider_platforms")
	}
	return &kolibri, nil
}

func (db *DB) ToggleSubLibraryFavorite(params models.OpenContentParams) (bool, error) {
	var activity models.OpenContentActivity
	if err := db.Model(&models.OpenContentActivity{}).Where("user_id = ? AND content_id = ?", params.UserID, params.ContentID).Order("request_ts DESC").First(&activity).Error; err != nil {
		log.Infof("activity %v", activity)
		return false, newNotFoundDBError(err, "open_content_activities")
	}
	var fav models.LibraryFavorite
	if err := db.Model(&models.LibraryFavorite{}).Where("user_id = ? AND content_id = ? AND open_content_url_id = ?", params.UserID, params.ContentID, activity.OpenContentUrlID).First(&fav).Error; err == nil {
		if err := db.Delete(&fav).Error; err != nil {
			return false, newNotFoundDBError(err, "library_favorites")
		}
	} else {
		newFav := models.LibraryFavorite{
			UserID:                params.UserID,
			ContentID:             params.ContentID,
			OpenContentUrlId:      activity.OpenContentUrlID,
			Name:                  params.Name,
			IsDisabled:            false,
			OpenContentProviderID: params.OpenContentProviderID,
		}
		if err := db.Create(&newFav).Error; err != nil {
			return false, newCreateDBError(err, "library_favorites")
		}
	}
	return true, nil
}

func (db *DB) ToggleLibraryFavorite(contentParams models.OpenContentParams) (bool, error) {
	var fav models.LibraryFavorite
	url := models.OpenContentUrl{}
	if db.Where("content_url = ?", contentParams.ContentUrl).First(&url).RowsAffected == 0 {
		url.ContentURL = contentParams.ContentUrl
		if err := db.Create(&url).Error; err != nil {
			log.Warn("unable to create content url for activity")
			return false, newCreateDBError(err, "open_content_urls")
		}
	}
	if err := db.Model(&models.LibraryFavorite{}).Where("user_id = ? AND content_id = ? AND open_content_url_id = ?", contentParams.UserID, contentParams.ContentID, url.ID).First(&fav).Error; err == nil {
		if err := db.Delete(&fav).Error; err != nil {
			return false, newNotFoundDBError(err, "library_favorites")
		}
	} else {
		newFav := models.LibraryFavorite{
			UserID:                contentParams.UserID,
			ContentID:             contentParams.ContentID,
			OpenContentUrlId:      url.ID,
			Name:                  contentParams.Name,
			OpenContentProviderID: contentParams.OpenContentProviderID,
		}
		if err := db.Create(&newFav).Error; err != nil {
			return false, newCreateDBError(err, "library_favorites")
		}
	}
	return true, nil
}

func (db *DB) GetUserFavorites(userID uint) ([]models.CombinedFavorite, error) {
	var openContentFavorites []models.CombinedFavorite
	if err := db.Debug().Table("library_favorites fav").
		Select(`
            fav.id,
            fav.name,
            'library' as type,
            fav.content_id,
            lib.image_url as thumbnail_url,
            ocp.description,
            NOT lib.visibility_status AS is_disabled,
            fav.open_content_provider_id,
            ocp.name AS provider_name
        `).
		Joins(`JOIN open_content_urls ocu ON ocu.id = fav.open_content_url_id`).
		Joins(`JOIN open_content_providers ocp ON ocp.id = fav.open_content_provider_id`).
		Joins(`JOIN libraries lib ON lib.id = fav.content_id`).
		Where("fav.user_id = ? AND fav.deleted_at IS NULL", userID).
		Scan(&openContentFavorites).Error; err != nil {
		return nil, newGetRecordsDBError(err, "library_favorites")
	}

	var videoFavorites []models.CombinedFavorite
	if err := db.Table("video_favorites vf").
		Select(`
            vf.id,
            videos.title as name,
            'video' as type,
            vf.video_id as content_id,
            videos.thumbnail_url,
            videos.description,
            videos.open_content_provider_id,
            videos.channel_title,
            -- Map visibility_status to is_disabled
            NOT videos.visibility_status as is_disabled
        `).
		Joins("left join videos on vf.video_id = videos.id").
		Where("vf.user_id = ? AND vf.deleted_at IS NULL", userID).
		Scan(&videoFavorites).Error; err != nil {
		return nil, newGetRecordsDBError(err, "video_favorites")
	}

	allFavorites := append(openContentFavorites, videoFavorites...)
	return allFavorites, nil
}
