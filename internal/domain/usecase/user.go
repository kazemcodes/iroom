package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	sanitize "github.com/iroom/iroom/internal/pkg/sanitize"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type UserUseCase struct {
	userRepo  *repository.UserRepo
	classRepo *repository.ClassRepo
	hasher    interface {
		Hash(password string) (string, error)
		Check(password, hash string) bool
	}
}

func NewUserUseCase(userRepo *repository.UserRepo, classRepo *repository.ClassRepo, hasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, classRepo: classRepo, hasher: hasher}
}

func (uc *UserUseCase) List(page, perPage int, search string) ([]entity.User, int64, error) {
	return uc.userRepo.List(page, perPage, search)
}

func (uc *UserUseCase) GetByID(id int64) (*entity.User, error) {
	return uc.userRepo.GetByID(id)
}

func (uc *UserUseCase) Create(email, password, displayName, phone, role string, callerRole string) (*entity.User, error) {
	existing, _ := uc.userRepo.GetByEmail(email)
	if existing != nil {
		return nil, fmt.Errorf("ایمیل قبلاً ثبت شده است")
	}

	hashedPassword, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("خطای داخلی")
	}

	user := &entity.User{
		Email:        email,
		PasswordHash: hashedPassword,
		DisplayName:  sanitize.Sanitize(displayName),
		Role:         entity.UserRole(role),
		Phone:        phone,
		IsActive:     true,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد کاربر")
	}
	return user, nil
}

func (uc *UserUseCase) Update(id int64, displayName, phone, role string, isActive *bool, callerRole string) error {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("کاربر یافت نشد")
	}

	if displayName != "" {
		user.DisplayName = sanitize.Sanitize(displayName)
	}
	if phone != "" {
		user.Phone = phone
	}
	if role != "" {
		user.Role = entity.UserRole(role)
	}
	if isActive != nil {
		user.IsActive = *isActive
	}
	return uc.userRepo.Update(user)
}

func (uc *UserUseCase) Delete(id int64) error {
	_, err := uc.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("کاربر یافت نشد")
	}
	return uc.userRepo.Delete(id)
}

func (uc *UserUseCase) BatchDelete(ids []int64) (int, int) {
	success, failure := 0, 0
	for _, id := range ids {
		user, err := uc.userRepo.GetByID(id)
		if err != nil || user.Role == entity.RoleAdmin {
			failure++
			continue
		}
		user.IsActive = false
		if err := uc.userRepo.Update(user); err != nil {
			failure++
		} else {
			success++
		}
	}
	return success, failure
}

func (uc *UserUseCase) Count() (int64, error) {
	return uc.userRepo.Count()
}

func (uc *UserUseCase) GetUserRooms(userID int64) ([]entity.Class, error) {
	return uc.classRepo.GetByUserID(userID)
}

func (uc *UserUseCase) ResetPassword(userID int64, newPassword string) error {
	hashedPassword, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("خطای داخلی")
	}
	return uc.userRepo.UpdatePassword(userID, hashedPassword)
}

func (uc *UserUseCase) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("کاربر یافت نشد")
	}

	if !uc.hasher.Check(oldPassword, user.PasswordHash) {
		return fmt.Errorf("رمز عبور فعلی اشتباه است")
	}

	hashedPassword, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("خطای داخلی")
	}

	return uc.userRepo.UpdatePassword(userID, hashedPassword)
}
