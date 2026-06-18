package entity

import "time"

// Class represents a virtual classroom owned by a teacher.
// Classes contain sessions (meetings) and enrolled students.
//
// Business rules:
//   - Each class is owned by exactly one teacher (teacherID)
//   - Students enroll via invite code or direct assignment
//   - max_students limits concurrent enrollment
//   - invite_code is a unique URL-friendly string for student self-enrollment
//   - isArchived=true hides the class from active listings
type Class struct {
	ID          int64     `json:"id" db:"id"`
	TeacherID   int64     `json:"teacher_id" db:"teacher_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	MaxStudents int       `json:"max_students" db:"max_students"`
	InviteCode  string    `json:"invite_code,omitempty" db:"invite_code"`
	Slug        string    `json:"slug,omitempty" db:"slug"` // URL-friendly name for /ch-{org}/{slug}/
	IsArchived  bool      `json:"is_archived" db:"is_archived"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ClassStudent maps students to classes with access levels.
// Represents the many-to-many relationship between users and classes.
//
// Access levels (matching Skyroom API):
//   1 = Regular student
//   2 = Presenter (can share screen/whiteboard)
//   3 = Operator (full control within the class)
type ClassStudent struct {
	ClassID   int64 `json:"class_id" db:"class_id"`
	StudentID int64 `json:"student_id" db:"student_id"`
	Access    int   `json:"access" db:"access"` // 1=student, 2=presenter, 3=operator
}
