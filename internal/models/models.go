package models

import (
	"time"

	"gorm.io/gorm"
)

// Role definitions
const (
	RoleAdmin   = "admin"
	RoleRelawan = "relawan"
	RoleKorban  = "korban"
)

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Email         string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash  string         `gorm:"size:255;not null" json:"-"`
	PlainPassword string         `gorm:"size:255" json:"-"`
	Role          string         `gorm:"type:enum('admin','relawan','korban');default:'korban';not null" json:"role"`
	Phone         string         `gorm:"size:20" json:"phone"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// Disaster Types
const (
	DisasterGempa     = "gempa"
	DisasterBanjir    = "banjir"
	DisasterLongsor   = "longsor"
	DisasterTsunami   = "tsunami"
	DisasterKebakaran = "kebakaran"
	DisasterLainnya   = "lainnya"
)

type Disaster struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:150;not null" json:"name"`
	DisasterType string         `gorm:"type:enum('gempa','banjir','longsor','tsunami','kebakaran','lainnya');not null" json:"disaster_type"`
	Location     string         `gorm:"size:255;not null" json:"location"`
	Date         time.Time      `json:"date"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Report Status
const (
	StatusPending   = "pending"
	StatusProcess   = "process"
	StatusCompleted = "completed"
)

// Urgency Levels
const (
	UrgencyLow      = "low"
	UrgencyMedium   = "medium"
	UrgencyHigh     = "high"
	UrgencyCritical = "critical"
)

type Report struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       *uint          `json:"user_id"`
	User         User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"user"`
	ReporterName string         `gorm:"size:100" json:"reporter_name"`
	Needs        string         `gorm:"type:text" json:"needs"`
	Latitude     float64        `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude    float64        `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Description  string         `gorm:"type:text" json:"description"`
	Status       string         `gorm:"type:enum('pending','process','completed');default:'pending';not null" json:"status"`
	UrgencyLevel string         `gorm:"type:enum('low','medium','high','critical');default:'medium'" json:"urgency_level"`
	ExtractedData string        `gorm:"type:text" json:"extracted_data"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Logistic struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ItemName  string         `gorm:"size:100;not null" json:"item_name"`
	Quantity  int            `gorm:"not null;default:0" json:"quantity"`
	Unit      string         `gorm:"size:50;not null" json:"unit"` // e.g., "kg", "dus", "liter"
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Distribution Status
const (
	DistStatusDelivering = "delivering" // relawan dalam perjalanan
	DistStatusDelivered  = "delivered"  // relawan sudah sampai & upload foto
	DistStatusVerified   = "verified"   // admin konfirmasi
	DistStatusRejected   = "rejected"   // admin tolak
)

type Distribution struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ReportID     uint           `gorm:"not null" json:"report_id"`
	Report       Report         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"report"`
	LogisticID   uint           `gorm:"not null" json:"logistic_id"`
	Logistic     Logistic       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"logistic"`
	QuantitySent int            `gorm:"not null" json:"quantity_sent"`
	Timestamp    time.Time      `gorm:"not null;autoCreateTime" json:"timestamp"`

	// Volunteer tracking fields
	VolunteerID    *uint      `json:"volunteer_id"`
	Volunteer      User       `gorm:"foreignKey:VolunteerID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"volunteer"`
	Status         string     `gorm:"size:20;default:'delivering'" json:"status"`
	ProofPhotoPath string     `gorm:"size:500" json:"proof_photo_path"`
	VolunteerLat   float64    `gorm:"type:decimal(10,8)" json:"volunteer_lat"`
	VolunteerLng   float64    `gorm:"type:decimal(11,8)" json:"volunteer_lng"`

	// Admin verification fields
	VerifiedByID   *uint      `json:"verified_by_id"`
	VerifiedBy     User       `gorm:"foreignKey:VerifiedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"verified_by"`
	VerifiedAt     *time.Time `json:"verified_at"`
	AdminNotes     string     `gorm:"type:text" json:"admin_notes"`

	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
