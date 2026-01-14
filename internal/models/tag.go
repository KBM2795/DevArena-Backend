package models

import "time"

// Tag represents a challenge category/tag
type Tag struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name      string    `json:"name" gorm:"uniqueIndex;type:varchar(100);not null"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;type:varchar(100);not null"`
	Category  string    `json:"category" gorm:"type:varchar(100)"` // e.g., "Language", "Topic", "Skill"
	Color     string    `json:"color" gorm:"type:varchar(20)"`     // For UI display
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Challenges []Challenge `json:"challenges,omitempty" gorm:"many2many:challenge_tags"`
}

// Common tag categories
const (
	TagCategoryLanguage = "Language"
	TagCategoryTopic    = "Topic"
	TagCategorySkill    = "Skill"
)
