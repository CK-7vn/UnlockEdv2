package models

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type OpenContentProvider struct {
	DatabaseFields
	Name             string `gorm:"size:255"  json:"name"`
	BaseUrl          string `gorm:"size:255;not null" json:"base_url"`
	Thumbnail        string `json:"thumbnail_url"`
	CurrentlyEnabled bool   `json:"currently_enabled"`
	Description      string `json:"description"`

	Videos []Video        `gorm:"foreignKey:OpenContentProviderID" json:"-"`
	Tasks  []RunnableTask `gorm:"foreignKey:OpenContentProviderID" json:"-"`
}

type OpenContentActivity struct {
	OpenContentProviderID uint      `gorm:"not null" json:"open_content_provider_id"`
	FacilityID            uint      `gorm:"not null" json:"facility_id"`
	UserID                uint      `gorm:"not null" json:"user_id"`
	ContentID             uint      `gorm:"not null" json:"content_id"`
	OpenContentUrlID      uint      `gorm:"not null" json:"open_content_url_id"`
	RequestTS             time.Time `gorm:"type:timestamp(0);default:CURRENT_TIMESTAMP" json:"request_ts"`

	User                *User                `gorm:"foreignKey:UserID" json:"-"`
	OpenContentProvider *OpenContentProvider `gorm:"foreignKey:OpenContentProviderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"open_content_provider"`
	Facility            *Facility            `json:"-" gorm:"foreignKey:FacilityID;references:ID"`
}

func (OpenContentActivity) TableName() string { return "open_content_activities" }

type LibraryFavorite struct {
	DatabaseFields
	UserID                uint   `gorm:"not null" json:"user_id"`
	ContentID             uint   `gorm:"not null" json:"content_id"`
	OpenContentUrlId      uint   `gorm:"not null" json:"open_content_url_id"`
	Name                  string `gorm:"size:255;not null" json:"name"`
	IsDisabled            bool   `gorm:"default:false;not null" json:"is_disabled"`
	OpenContentProviderID uint   `gorm:"not null" json:"open_content_provider_id"`
	User                  *User  `json:"user" gorm:"foreignKey:UserID"`
}

type CombinedFavorite struct {
	ID                    uint   `json:"id"`
	ContentID             uint   `json:"content_id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	ThumbnailUrl          string `json:"thumbnail_url"`
	Description           string `json:"description"`
	IsDisabled            bool   `json:"is_disabled"`
	OpenContentProviderID uint   `json:"open_content_provider_id"`
	ProviderName          string `json:"provider_name,omitempty"`
	ChannelTitle          string `json:"channel_title,omitempty"`
}

func (LibraryFavorite) TableName() string { return "library_favorites" }

type OpenContentUrl struct {
	ID         uint   `gorm:"primaryKey" json:"-"`
	ContentURL string `gorm:"size:255" json:"content_url"`
}

type OpenContentParams struct {
	UserID                uint
	ContentID             uint
	Name                  string
	ContentUrl            string
	OpenContentProviderID uint
}

const (
	KolibriThumbnailUrl string = "https://learningequality.org/static/assets/kolibri-ecosystem-logos/blob-logo.svg"
	Kiwix               string = "Kiwix"
	KolibriDescription  string = "Kolibri provides an extensive library of educational content suitable for all learning levels."
	KiwixThumbnailURL   string = "/kiwix.jpg"
	KiwixDescription    string = "Kiwix is an offline reader that allows you to host a wide array of educational content."
	KiwixLibraryUrl     string = "https://library.kiwix.org"
	YoutubeThumbnail    string = "/youtube.png"
	Youtube             string = "Youtube"
	YoutubeApi          string = "https://www.googleapis.com/youtube/v3/videos"
	YoutubeDescription  string = "Hand pick videos to be available to students from youtube URL's"
)

func (cp *OpenContentProvider) BeforeCreate(tx *gorm.DB) error {
	if cp.Name == Youtube && cp.BaseUrl == "" {
		cp.BaseUrl = YoutubeApi
	}
	if cp.BaseUrl != "" && !strings.HasPrefix(cp.BaseUrl, "http") {
		cp.BaseUrl = fmt.Sprintf("https://%s", cp.BaseUrl)
	}
	return nil
}

func (OpenContentProvider) TableName() string { return "open_content_providers" }
