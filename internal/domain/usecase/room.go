package usecase

import (
	"fmt"
	"time"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/errors"
	"github.com/iroom/iroom/internal/pkg/slug"
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

func (uc *RoomUseCase) Create(ownerID int64, name, description, color string, maxUsers int, inviteCode string) (*entity.Room, error) {
	room := &entity.Room{
		OwnerID:           ownerID,
		Name:              name,
		Description:       description,
		Color:             color,
		Slug:              slug.Generate(name),
		GuestLoginEnabled: true,
		MaxUsers:          maxUsers,
		InviteCode:        inviteCode,
	}
	if err := uc.roomRepo.Create(room); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}
	return room, nil
}

func (uc *RoomUseCase) GetByID(id int64) (*entity.Room, error) {
	room, err := uc.roomRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	return room, nil
}

func (uc *RoomUseCase) GetBySlug(slug string) (*entity.Room, error) {
	room, err := uc.roomRepo.GetBySlug(slug)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	return room, nil
}

func (uc *RoomUseCase) List(page, perPage int, search string) ([]entity.Room, int64, error) {
	return uc.roomRepo.ListAll(page, perPage, search)
}

func (uc *RoomUseCase) Update(id, actorID int64, role, name, description, color string, guestLoginEnabled bool, maxUsers int, inviteCode string) (*entity.Room, error) {
	room, err := uc.roomRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return nil, errors.ErrForbidden
	}
	if name != "" {
		room.Name = name
		room.Slug = slug.Generate(name)
	}
	if description != "" {
		room.Description = description
	}
	if color != "" {
		room.Color = color
	}
	room.GuestLoginEnabled = guestLoginEnabled
	if maxUsers > 0 {
		room.MaxUsers = maxUsers
	}
	if inviteCode != "" {
		room.InviteCode = inviteCode
	}
	if err := uc.roomRepo.Update(room); err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}
	return room, nil
}

func (uc *RoomUseCase) Delete(id, actorID int64, role string) error {
	room, err := uc.roomRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return errors.ErrForbidden
	}
	return uc.roomRepo.Delete(id)
}

func (uc *RoomUseCase) AddUser(roomID, userID, actorID int64, role string, access int, actorRole string) error {
	if role == "" {
		role = "student"
	}
	if access < 1 {
		access = 1
	}
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && actorRole != "admin" {
		return errors.ErrForbidden
	}
	return uc.roomRepo.AddUser(roomID, userID, role, access)
}

func (uc *RoomUseCase) UpdateSettings(roomID, actorID int64, actorRole string, settings *entity.RoomSettings) error {
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && actorRole != "admin" {
		return errors.ErrForbidden
	}
	return uc.roomRepo.UpdateSettings(settings)
}

func (uc *RoomUseCase) RemoveUser(roomID, userID, actorID int64, role string) error {
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return errors.ErrForbidden
	}
	return uc.roomRepo.RemoveUser(roomID, userID)
}

func (uc *RoomUseCase) UpdateUserAccess(roomID, userID, actorID int64, role string, access int) error {
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return errors.ErrForbidden
	}
	return uc.roomRepo.UpdateUserAccess(roomID, userID, access)
}

func (uc *RoomUseCase) GetUsers(roomID int64) ([]entity.User, error) {
	users, err := uc.roomRepo.GetUsers(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
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

func (uc *RoomUseCase) RegenerateCode(roomID, actorID int64, role string) (string, error) {
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return "", errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return "", errors.ErrForbidden
	}
	code := fmt.Sprintf("%d-%d", roomID, time.Now().UnixMilli())
	if err := uc.roomRepo.UpdateInviteCode(roomID, code); err != nil {
		return "", fmt.Errorf("failed to regenerate code: %w", err)
	}
	return code, nil
}

func (uc *RoomUseCase) JoinByCode(code string) (*entity.Room, error) {
	room, err := uc.roomRepo.GetByInviteCode(code)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	return room, nil
}

func (uc *RoomUseCase) GetUserRooms(userID int64) ([]entity.Room, error) {
	rooms, err := uc.roomRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}
	if rooms == nil {
		rooms = []entity.Room{}
	}
	return rooms, nil
}
