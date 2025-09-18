package models

import "time"

type Cat struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name" validate:"required,min=2"`
	YearsOfExperience int       `json:"years_of_experience" validate:"gte=0"`
	Breed             string    `json:"breed" validate:"required"`
	SalaryCents       int64     `json:"salary_cents" validate:"gte=0"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Mission struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	AssignedCatID *uint     `json:"assigned_cat_id"`
	Completed     bool      `json:"completed"`
	Targets       []Target  `json:"targets" gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Target struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MissionID uint      `json:"mission_id"`
	Name      string    `json:"name" validate:"required,min=2"`
	Country   string    `json:"country" validate:"required"`
	Notes     string    `json:"notes"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Breed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
