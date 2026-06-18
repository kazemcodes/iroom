package entity

import "time"

type Class struct {
	ID          int64     `json:"id"`
	TeacherID   int64     `json:"teacher_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	MaxStudents int       `json:"max_students"`
	InviteCode  string    `json:"invite_code,omitempty"`
	IsArchived  bool      `json:"is_archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ClassStudent struct {
	ClassID   int64 `json:"class_id"`
	StudentID int64 `json:"student_id"`
	Access    int   `json:"access"`
}
