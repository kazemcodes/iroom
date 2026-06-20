package usecase

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type RoomUseCase struct {
	roomRepo    *repository.RoomRepo
	userRepo    *repository.UserRepo
	sessionRepo *repository.SessionRepo
}

func NewRoomUseCase(roomRepo *repository.RoomRepo, userRepo *repository.UserRepo, sessionRepo *repository.SessionRepo) *RoomUseCase {
	return &RoomUseCase{roomRepo: roomRepo, userRepo: userRepo, sessionRepo: sessionRepo}
}

func generateRoomSlug(name string) string {
	slug := strings.ToLower(name)
	replacements := map[string]string{
		" ": "-", "‌": "", "۰": "0", "۱": "1", "۲": "2", "۳": "3",
		"۴": "4", "۵": "5", "۶": "6", "۷": "7", "۸": "8", "۹": "9",
	}
	for k, v := range replacements {
		slug = strings.ReplaceAll(slug, k, v)
	}
	var result []rune
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result = append(result, r)
		}
	}
	slug = string(result)
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = fmt.Sprintf("room-%d", time.Now().UnixMilli())
	}
	return slug
}

func (uc *RoomUseCase) Create(ownerID int64, name, description, color string) (*entity.Room, error) {
	slug := generateRoomSlug(name)
	room := &entity.Room{
		OwnerID:           ownerID,
		Name:              name,
		Description:       description,
		Color:             color,
		Slug:              slug,
		GuestLoginEnabled: true,
	}
	if err := uc.roomRepo.Create(room); err != nil {
		return nil, fmt.Errorf("خطا در ایجاد اتاق")
	}
	return room, nil
}

func (uc *RoomUseCase) GetByID(id int64) (*entity.Room, error) {
	return uc.roomRepo.GetByID(id)
}

func (uc *RoomUseCase) GetBySlug(slug string) (*entity.Room, error) {
	return uc.roomRepo.GetBySlug(slug)
}

func (uc *RoomUseCase) List(page, perPage int, search string) ([]entity.Room, int64, error) {
	return uc.roomRepo.ListAll(page, perPage, search)
}

func (uc *RoomUseCase) Update(id int64, name, description, color string, guestLoginEnabled bool) (*entity.Room, error) {
	room, err := uc.roomRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("اتاق یافت نشد")
	}
	if name != "" {
		room.Name = name
		room.Slug = generateRoomSlug(name)
	}
	if description != "" {
		room.Description = description
	}
	if color != "" {
		room.Color = color
	}
	room.GuestLoginEnabled = guestLoginEnabled
	if err := uc.roomRepo.Update(room); err != nil {
		return nil, fmt.Errorf("خطا در بروزرسانی")
	}
	return room, nil
}

func (uc *RoomUseCase) Delete(id int64) error {
	return uc.roomRepo.Delete(id)
}

func (uc *RoomUseCase) AddUser(roomID, userID int64, role string) error {
	if role == "" {
		role = "student"
	}
	return uc.roomRepo.AddUser(roomID, userID, role)
}

func (uc *RoomUseCase) RemoveUser(roomID, userID int64) error {
	return uc.roomRepo.RemoveUser(roomID, userID)
}

func (uc *RoomUseCase) GetUsers(roomID int64) ([]entity.User, error) {
	users, err := uc.roomRepo.GetUsers(roomID)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت کاربران")
	}
	if users == nil {
		users = []entity.User{}
	}
	return users, nil
}

func (uc *RoomUseCase) GetUserCount(roomID int64) (int, error) {
	return uc.roomRepo.GetUserCount(roomID)
}

func (uc *RoomUseCase) GetSettings(roomID int64) (*entity.RoomSettings, error) {
	return uc.roomRepo.GetSettings(roomID)
}

func (uc *RoomUseCase) UpdateSettings(roomID int64, settings *entity.RoomSettings) error {
	return uc.roomRepo.UpdateSettings(settings)
}

func (uc *RoomUseCase) GetActiveSessionCount(roomID int64) (int, error) {
	sessions, err := uc.sessionRepo.ListByClass(roomID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, s := range sessions {
		if s.Status == entity.SessionLive {
			count++
		}
	}
	return count, nil
}
