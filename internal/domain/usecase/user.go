package usecase

import (
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
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

func (uc *UserUseCase) Create(email, password, displayName, phone, role string) (*entity.User, error) {
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
		DisplayName:  displayName,
		Role:         entity.UserRole(role),
		Phone:        phone,
		IsActive:     true,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد کاربر")
	}
	return user, nil
}

func (uc *UserUseCase) Update(id int64, displayName, phone, role string, isActive *bool) error {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("کاربر یافت نشد")
	}
	if displayName != "" {
		user.DisplayName = displayName
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
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("کاربر یافت نشد")
	}
	user.IsActive = false
	return uc.userRepo.Update(user)
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
