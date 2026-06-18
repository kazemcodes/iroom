package usecase

import (
	"fmt"
	"time"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

// ClassUseCase handles all class-related business logic.
// Classes are virtual classrooms owned by teachers, containing enrolled students and sessions.
type ClassUseCase struct {
	classRepo   *repository.ClassRepo
	sessionRepo *repository.SessionRepo
}

func NewClassUseCase(classRepo *repository.ClassRepo, sessionRepo *repository.SessionRepo) *ClassUseCase {
	return &ClassUseCase{classRepo: classRepo, sessionRepo: sessionRepo}
}

func (uc *ClassUseCase) Create(teacherID int64, name, description, color string, maxStudents int) (*entity.Class, error) {
	c := &entity.Class{
		TeacherID:   teacherID,
		Name:        name,
		Description: description,
		Color:       color,
		MaxStudents: maxStudents,
	}
	if err := uc.classRepo.Create(c); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد کلاس")
	}
	return c, nil
}

func (uc *ClassUseCase) GetByID(id int64) (*entity.Class, error) {
	c, err := uc.classRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("کلاس یافت نشد")
	}
	return c, nil
}

func (uc *ClassUseCase) List(teacherID int64, page, perPage int, search string) ([]entity.Class, int64, error) {
	if teacherID > 0 {
		classes, err := uc.classRepo.ListByTeacher(teacherID)
		return classes, int64(len(classes)), err
	}
	return uc.classRepo.ListAll(page, perPage, search)
}

func (uc *ClassUseCase) Update(id, teacherID int64, name, description, color string, maxStudents int, role string) (*entity.Class, error) {
	c, err := uc.classRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("کلاس یافت نشد")
	}
	if c.TeacherID != teacherID && role != "admin" {
		return nil, fmt.Errorf("شما اجازه ویرایش این کلاس را ندارید")
	}

	if name != "" {
		c.Name = name
	}
	if description != "" {
		c.Description = description
	}
	if color != "" {
		c.Color = color
	}
	if maxStudents > 0 {
		c.MaxStudents = maxStudents
	}

	if err := uc.classRepo.Update(c); err != nil {
		return nil, fmt.Errorf("خطا در بروزرسانی")
	}
	return c, nil
}

func (uc *ClassUseCase) Delete(id, teacherID int64, role string) error {
	c, err := uc.classRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("کلاس یافت نشد")
	}
	if c.TeacherID != teacherID && role != "admin" {
		return fmt.Errorf("شما اجازه حذف این کلاس را ندارید")
	}
	return uc.classRepo.Delete(id)
}

func (uc *ClassUseCase) Enroll(classID, studentID int64) error {
	return uc.classRepo.Enroll(classID, studentID)
}

func (uc *ClassUseCase) RemoveUser(classID, userID, actorID int64, role string) error {
	c, err := uc.classRepo.GetByID(classID)
	if err != nil {
		return fmt.Errorf("کلاس یافت نشد")
	}
	if c.TeacherID != actorID && role != "admin" {
		return fmt.Errorf("شما اجازه حذف کاربر از این کلاس را ندارید")
	}
	return uc.classRepo.RemoveStudent(classID, userID)
}

func (uc *ClassUseCase) UpdateUserAccess(classID, userID, actorID int64, role string, access int) error {
	c, err := uc.classRepo.GetByID(classID)
	if err != nil {
		return fmt.Errorf("کلاس یافت نشد")
	}
	if c.TeacherID != actorID && role != "admin" {
		return fmt.Errorf("شما اجازه تغییر دسترسی در این کلاس را ندارید")
	}
	return uc.classRepo.UpdateStudentAccess(classID, userID, access)
}

func (uc *ClassUseCase) GetStudents(classID int64) ([]entity.User, error) {
	students, err := uc.classRepo.GetStudents(classID)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت دانش‌آموزان")
	}
	if students == nil {
		students = []entity.User{}
	}
	return students, nil
}

func (uc *ClassUseCase) GetURL(id int64) string {
	return fmt.Sprintf("/classroom/join/%d", id)
}

func (uc *ClassUseCase) GetUserRooms(userID int64) ([]entity.Class, error) {
	rooms, err := uc.classRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت اتاق‌ها")
	}
	if rooms == nil {
		rooms = []entity.Class{}
	}
	return rooms, nil
}

func (uc *ClassUseCase) JoinByCode(code string) (*entity.Class, error) {
	c, err := uc.classRepo.GetByInviteCode(code)
	if err != nil {
		return nil, fmt.Errorf("کلاس یافت نشد")
	}
	return c, nil
}

func (uc *ClassUseCase) RegenerateCode(classID, teacherID int64, role string) (string, error) {
	c, err := uc.classRepo.GetByID(classID)
	if err != nil {
		return "", fmt.Errorf("کلاس یافت نشد")
	}
	if c.TeacherID != teacherID && role != "admin" {
		return "", fmt.Errorf("شما اجازه تغییر این کلاس را ندارید")
	}

	code := fmt.Sprintf("%d-%d", classID, time.Now().UnixMilli())
	if err := uc.classRepo.UpdateInviteCode(classID, code); err != nil {
		return "", fmt.Errorf("خطا در بروزرسانی کد")
	}
	return code, nil
}
