package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type ClassRepo struct {
	db *sql.DB
}

func NewClassRepo(db *sql.DB) *ClassRepo {
	return &ClassRepo{db: db}
}

func (r *ClassRepo) Create(c *models.Class) error {
	result, err := r.db.Exec(
		`INSERT INTO classes (teacher_id, name, description, color, max_students) VALUES (?, ?, ?, ?, ?)`,
		c.TeacherID, c.Name, c.Description, c.Color, c.MaxStudents,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}

func (r *ClassRepo) GetByID(id int64) (*models.Class, error) {
	c := &models.Class{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, name, description, color, max_students, created_at, updated_at FROM classes WHERE id = ?`, id,
	).Scan(&c.ID, &c.TeacherID, &c.Name, &c.Description, &c.Color, &c.MaxStudents, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ClassRepo) ListByTeacher(teacherID int64) ([]models.Class, error) {
	rows, err := r.db.Query(
		`SELECT id, teacher_id, name, description, color, max_students, created_at, updated_at FROM classes WHERE teacher_id = ? ORDER BY id DESC`, teacherID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.TeacherID, &c.Name, &c.Description, &c.Color, &c.MaxStudents, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}

func (r *ClassRepo) ListByStudent(studentID int64) ([]models.Class, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.teacher_id, c.name, c.description, c.color, c.max_students, c.created_at, c.updated_at
		 FROM classes c JOIN class_students cs ON c.id = cs.class_id WHERE cs.student_id = ? ORDER BY c.id DESC`, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.TeacherID, &c.Name, &c.Description, &c.Color, &c.MaxStudents, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}

func (r *ClassRepo) ListAll(page, perPage int, search string) ([]models.Class, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM classes WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		countQuery += ` AND (name LIKE ? OR description LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s)
	}

	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `SELECT id, teacher_id, name, description, color, max_students, created_at, updated_at FROM classes WHERE 1=1`
	if search != "" {
		query += ` AND (name LIKE ? OR description LIKE ?)`
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, perPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.TeacherID, &c.Name, &c.Description, &c.Color, &c.MaxStudents, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		classes = append(classes, c)
	}
	return classes, total, nil
}

func (r *ClassRepo) Enroll(classID, studentID int64) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO class_students (class_id, student_id) VALUES (?, ?)`,
		classID, studentID,
	)
	return err
}

func (r *ClassRepo) GetStudents(classID int64) ([]models.User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.email, u.password_hash, u.display_name, u.role, u.phone, u.is_active, u.created_at, u.updated_at
		 FROM users u JOIN class_students cs ON u.id = cs.student_id WHERE cs.class_id = ?`, classID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *ClassRepo) IsEnrolled(classID, studentID int64) bool {
	var count int
	r.db.QueryRow(`SELECT COUNT(*) FROM class_students WHERE class_id = ? AND student_id = ?`, classID, studentID).Scan(&count)
	return count > 0
}

func (r *ClassRepo) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM classes`).Scan(&count)
	return count, err
}

func (r *ClassRepo) Update(c *models.Class) error {
	_, err := r.db.Exec(
		`UPDATE classes SET name = ?, description = ?, color = ?, max_students = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		c.Name, c.Description, c.Color, c.MaxStudents, c.ID,
	)
	return err
}

func (r *ClassRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM classes WHERE id = ?`, id)
	return err
}

// Invite code methods

func (r *ClassRepo) UpdateInviteCode(classID int64, code string) error {
	_, err := r.db.Exec(
		`UPDATE classes SET invite_code = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		code, classID,
	)
	return err
}

func (r *ClassRepo) GetByInviteCode(code string) (*models.Class, error) {
	c := &models.Class{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, name, description, color, max_students, invite_code, is_archived, created_at, updated_at FROM classes WHERE invite_code = ?`, code,
	).Scan(&c.ID, &c.TeacherID, &c.Name, &c.Description, &c.Color, &c.MaxStudents, &c.InviteCode, &c.IsArchived, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ClassRepo) GetByIDWithInvite(id int64) (*models.Class, error) {
	c := &models.Class{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, name, description, color, max_students, invite_code, is_archived, created_at, updated_at FROM classes WHERE id = ?`, id,
	).Scan(&c.ID, &c.TeacherID, &c.Name, &c.Description, &c.Color, &c.MaxStudents, &c.InviteCode, &c.IsArchived, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ClassRepo) RemoveStudent(classID, studentID int64) error {
	_, err := r.db.Exec(`DELETE FROM class_students WHERE class_id = ? AND student_id = ?`, classID, studentID)
	return err
}

func (r *ClassRepo) UpdateStudentAccess(classID, studentID int64, access int) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO class_students (class_id, student_id) VALUES (?, ?)`,
		classID, studentID,
	)
	return err
}

func (r *ClassRepo) GetByUserID(userID int64) ([]models.Class, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.teacher_id, c.name, c.description, c.color, c.max_students, c.created_at, c.updated_at
		 FROM classes c JOIN class_students cs ON c.id = cs.class_id WHERE cs.student_id = ?
		 UNION
		 SELECT c.id, c.teacher_id, c.name, c.description, c.color, c.max_students, c.created_at, c.updated_at
		 FROM classes c WHERE c.teacher_id = ?`, userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.TeacherID, &c.Name, &c.Description, &c.Color, &c.MaxStudents, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}
