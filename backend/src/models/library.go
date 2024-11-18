package models

type Library struct {
	DatabaseFields
	OpenContentProviderID uint    `gorm:"not null" json:"open_content_provider_id"`
	ExternalID            *string `json:"external_id"`
	Name                  string  `gorm:"size:255;not null" json:"name"`
	Language              *string `gorm:"size:255" json:"language"`
	Description           *string `json:"description"`
	Path                  string  `gorm:"not null" json:"url"`
	ImageUrl              *string `json:"image_url"`
	VisibilityStatus      bool    `gorm:"default:false;not null" json:"visibility_status"`

	OpenContentProvider *OpenContentProvider `gorm:"foreignKey:OpenContentProviderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"open_content_provider"`
}

func (Library) TableName() string { return "libraries" }

type LibraryProxyPO struct {
	ID                    uint
	OpenContentProviderID uint
	Path                  string
	BaseUrl               string
	VisibilityStatus      bool
}

type LibraryDto struct {
	ID               uint    `json:"id"`
	ExternalID       *string `json:"external_id"`
	Name             string  `json:"name"`
	Language         *string `json:"language"`
	Description      *string `json:"description"`
	Path             string  `json:"url"`
	ImageUrl         *string `json:"image_url"`
	VisibilityStatus bool    `json:"visibility_status"`
	IsFavorited      bool    `json:"is_favorited"`

	//open_content_provider
	OpenContentProviderID   uint   `json:"open_content_provider_id"`
	OpenContentProviderName string `json:"open_content_provider_name"`
	BaseUrl                 string `json:"base_url"`
	Thumbnail               string `json:"thumbnail_url"`
	CurrentlyEnabled        bool   `json:"currently_enabled"`
	OpenContentProviderDesc string `json:"open_content_provider_description"`
}
