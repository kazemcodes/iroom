package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/errors"
	"github.com/iroom/iroom/internal/pkg/slug"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

const (
	defaultMaxUsers = 50
	minMaxUsers     = 1
	maxMaxUsers     = 1000
	minSessionMin   = 5
	maxSessionMin   = 1440
)

var validRoomRoles = map[string]bool{"teacher": true, "student": true}

func isValidAccess(a int) bool { return a >= 1 && a <= 3 }

type RoomUseCase struct {
	roomRepo    *repository.RoomRepo
	userRepo    *repository.UserRepo
	sessionRepo *repository.SessionRepo
}

func NewRoomUseCase(roomRepo *repository.RoomRepo, userRepo *repository.UserRepo, sessionRepo *repository.SessionRepo) *RoomUseCase {
	return &RoomUseCase{roomRepo: roomRepo, userRepo: userRepo, sessionRepo: sessionRepo}
}

func (uc *RoomUseCase) Create(ownerID int64, name, description, color string, maxUsers int, inviteCode string) (*entity.Room, error) {
	if name == "" {
		return nil, errors.ErrValidation
	}
	if maxUsers <= 0 {
		maxUsers = defaultMaxUsers
	}
	if maxUsers > maxMaxUsers {
		maxUsers = maxMaxUsers
	}
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

// RoomUpdate captures partial-update fields. nil = leave unchanged.
type RoomUpdate struct {
	Name              *string
	Description       *string
	Color             *string
	GuestLoginEnabled *bool
	MaxUsers          *int
	InviteCode        *string
}

func (uc *RoomUseCase) Update(id, actorID int64, role string, u RoomUpdate) (*entity.Room, error) {
	room, err := uc.roomRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return nil, errors.ErrForbidden
	}
	if u.Name != nil {
		if *u.Name == "" {
			return nil, errors.ErrValidation
		}
		room.Name = *u.Name
		// slug intentionally NOT regenerated — immutable to preserve links
	}
	if u.Description != nil {
		room.Description = *u.Description
	}
	if u.Color != nil {
		room.Color = *u.Color
	}
	if u.GuestLoginEnabled != nil {
		room.GuestLoginEnabled = *u.GuestLoginEnabled
	}
	if u.MaxUsers != nil {
		mu := *u.MaxUsers
		if mu < minMaxUsers || mu > maxMaxUsers {
			return nil, errors.ErrValidation
		}
		room.MaxUsers = mu
	}
	if u.InviteCode != nil {
		room.InviteCode = *u.InviteCode
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
	if !validRoomRoles[role] {
		return errors.ErrValidation
	}
	if access == 0 {
		access = 1
	}
	if !isValidAccess(access) {
		return errors.ErrValidation
	}
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && actorRole != "admin" {
		return errors.ErrForbidden
	}
	// enforce MaxUsers cap
	count, _ := uc.roomRepo.GetUserCount(roomID)
	if !uc.roomRepo.IsUserInRoom(roomID, userID) && room.MaxUsers > 0 && count >= room.MaxUsers {
		return errors.ErrConflict
	}
	return uc.roomRepo.AddUser(roomID, userID, role, access)
}

func (uc *RoomUseCase) UpdateSettings(roomID, actorID int64, actorRole string, settings *entity.RoomSettings) error {
	if settings.MaxUsers < minMaxUsers || settings.MaxUsers > maxMaxUsers {
		return errors.ErrValidation
	}
	if settings.SessionAutoEndMinutes < minSessionMin || settings.SessionAutoEndMinutes > maxSessionMin {
		return errors.ErrValidation
	}
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && actorRole != "admin" {
		return errors.ErrForbidden
	}
	settings.RoomID = roomID
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
	if !isValidAccess(access) {
		return errors.ErrValidation
	}
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return errors.ErrNotFound
	}
	if room.OwnerID != actorID && role != "admin" {
		return errors.ErrForbidden
	}
	if !uc.roomRepo.IsUserInRoom(roomID, userID) {
		return errors.ErrNotFound
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
	// 8 bytes = 16 hex chars, ~64 bits entropy. unguessable, no ID leak.
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}
	code := hex.EncodeToString(buf)
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
