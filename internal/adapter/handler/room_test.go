package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRoomHandler(t *testing.T) (*RoomHandler, echo.Context) {
	t.Helper()
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := usecase.NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	h := NewRoomHandler(uc)
	return h, nil
}

func newEchoContext(method, path string, body interface{}, params map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	for k, v := range params {
		c.SetParamNames(k)
		c.SetParamValues(v)
	}

	return c, rec
}

func TestRoomHandler_Create(t *testing.T) {
	h, _ := setupRoomHandler(t)

	body := map[string]string{
		"name":        "Test Room",
		"description": "A test room",
		"color":       "#FF0000",
	}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", body, nil)
	c.Set("user_id", int64(1))

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Test Room", data["name"])
	assert.Equal(t, "test-room", data["slug"])
	assert.True(t, data["guest_login_enabled"].(bool))
}

func TestRoomHandler_Create_MissingName(t *testing.T) {
	h, _ := setupRoomHandler(t)

	body := map[string]string{
		"description": "No name",
	}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", body, nil)
	c.Set("user_id", int64(1))

	require.NoError(t, h.Create(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoomHandler_GetByID(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Get Room", "color": "#000"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/"+strconv.FormatInt(roomID, 10), nil, map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.GetByID(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRoomHandler_GetByID_InvalidID(t *testing.T) {
	h, _ := setupRoomHandler(t)

	c, rec := newEchoContext(http.MethodGet, "/api/rooms/abc", nil, map[string]string{"id": "abc"})
	require.NoError(t, h.GetByID(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoomHandler_GetBySlug(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Slug Room"}
	c, _ := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/slug/slug-room", nil, map[string]string{"slug": "slug-room"})
	require.NoError(t, h.GetBySlug(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRoomHandler_GetBySlug_NotFound(t *testing.T) {
	h, _ := setupRoomHandler(t)

	c, rec := newEchoContext(http.MethodGet, "/api/rooms/slug/doesnt-exist", nil, map[string]string{"slug": "doesnt-exist"})
	require.NoError(t, h.GetBySlug(c))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRoomHandler_List(t *testing.T) {
	h, _ := setupRoomHandler(t)

	for i := 0; i < 3; i++ {
		body := map[string]string{"name": "Room " + strconv.Itoa(i)}
		c, _ := newEchoContext(http.MethodPost, "/api/rooms", body, nil)
		c.Set("user_id", int64(1))
		require.NoError(t, h.Create(c))
	}

	c, rec := newEchoContext(http.MethodGet, "/api/rooms?page=1&per_page=10", nil, nil)
	require.NoError(t, h.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["per_page"])
}

func TestRoomHandler_List_DefaultPagination(t *testing.T) {
	h, _ := setupRoomHandler(t)

	c, rec := newEchoContext(http.MethodGet, "/api/rooms", nil, nil)
	require.NoError(t, h.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRoomHandler_Update(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Old Name", "color": "#000"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	guestLogin := false
	updateBody := map[string]interface{}{
		"name":                "New Name",
		"description":         "Updated",
		"color":               "#FFF",
		"guest_login_enabled": guestLogin,
	}
	c2, rec2 := newEchoContext(http.MethodPut, "/api/rooms/"+strconv.FormatInt(roomID, 10), updateBody, map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.Update(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)

	var updateResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &updateResp))
	updateData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, "New Name", updateData["name"])
	assert.Equal(t, false, updateData["guest_login_enabled"])
}

func TestRoomHandler_Update_InvalidID(t *testing.T) {
	h, _ := setupRoomHandler(t)

	c, rec := newEchoContext(http.MethodPut, "/api/rooms/abc", map[string]string{"name": "X"}, map[string]string{"id": "abc"})
	require.NoError(t, h.Update(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoomHandler_Delete(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "To Delete"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	c2, rec2 := newEchoContext(http.MethodDelete, "/api/rooms/"+strconv.FormatInt(roomID, 10), nil, map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.Delete(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRoomHandler_Delete_InvalidID(t *testing.T) {
	h, _ := setupRoomHandler(t)

	c, rec := newEchoContext(http.MethodDelete, "/api/rooms/abc", nil, map[string]string{"id": "abc"})
	require.NoError(t, h.Delete(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoomHandler_AddUser(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Room With User"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp)) //nolint:errcheck
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	addBody := map[string]interface{}{
		"user_id": 2,
		"role":    "teacher",
	}
	c2, rec2 := newEchoContext(http.MethodPost, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/users", addBody, map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.AddUser(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRoomHandler_RemoveUser(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Room"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	h.roomUC.AddUser(roomID, 5, "student")

	c2, rec2 := newEchoContext(http.MethodDelete, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/users/5", nil,
		map[string]string{"id": strconv.FormatInt(roomID, 10), "userId": "5"})
	require.NoError(t, h.RemoveUser(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRoomHandler_GetUsers(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Room"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/users", nil,
		map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.GetUsers(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRoomHandler_GetSettings(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Room"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/settings", nil,
		map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.GetSettings(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &resp))
	settings := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(50), settings["max_users"])
	assert.Equal(t, true, settings["recording_enabled"])
}

func TestRoomHandler_UpdateSettings(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Room"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["data"].(map[string]interface{})
	roomID := int64(data["id"].(float64))

	settingsBody := map[string]interface{}{
		"max_users":                  100,
		"recording_enabled":          false,
		"allow_student_video":        true,
		"allow_student_audio":        false,
		"allow_student_screen_share": true,
		"allow_student_whiteboard":   true,
		"allow_student_chat":         false,
		"session_auto_end_minutes":   60,
		"waiting_room_enabled":       true,
	}
	c2, rec2 := newEchoContext(http.MethodPut, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/settings", settingsBody,
		map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.UpdateSettings(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRoomHandler_GetInfo(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Info Room"}
	c, _ := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/info/info-room", nil, map[string]string{"slug": "info-room"})
	require.NoError(t, h.GetInfo(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &resp))
	info := resp["data"].(map[string]interface{})
	assert.NotNil(t, info["room"])
	assert.Equal(t, float64(0), info["user_count"])
	assert.Equal(t, float64(0), info["active_sessions"])
}

func TestRoomHandler_GetInfo_NotFound(t *testing.T) {
	h, _ := setupRoomHandler(t)

	c, rec := newEchoContext(http.MethodGet, "/api/rooms/info/missing", nil, map[string]string{"slug": "missing"})
	require.NoError(t, h.GetInfo(c))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRoomHandler_Create_ResponseHasAllFields(t *testing.T) {
	h, _ := setupRoomHandler(t)

	body := map[string]string{"name": "Full Room", "description": "Desc", "color": "#ABC"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", body, nil)
	c.Set("user_id", int64(42))
	require.NoError(t, h.Create(c))

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})

	assert.NotZero(t, data["id"], "id must be present")
	assert.Equal(t, float64(42), data["owner_id"], "owner_id must be present and correct")
	assert.Equal(t, "Full Room", data["name"])
	assert.Equal(t, "Desc", data["description"])
	assert.Equal(t, "#ABC", data["color"])
	assert.Equal(t, "full-room", data["slug"])
	assert.Equal(t, true, data["guest_login_enabled"])
	assert.NotEmpty(t, data["created_at"], "created_at must be present")
	assert.NotEmpty(t, data["updated_at"], "updated_at must be present")
}

func TestRoomHandler_GetByID_ResponseHasAllFields(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Get Fields", "description": "D", "color": "#111"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	roomID := int64(createResp["data"].(map[string]interface{})["id"].(float64))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/"+strconv.FormatInt(roomID, 10), nil, map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.GetByID(c2))

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})

	assert.NotZero(t, data["id"])
	assert.NotZero(t, data["owner_id"], "owner_id must be in response")
	assert.Equal(t, "Get Fields", data["name"])
	assert.Equal(t, "D", data["description"])
	assert.Equal(t, "#111", data["color"])
	assert.NotEmpty(t, data["slug"])
	assert.NotEmpty(t, data["created_at"])
	assert.NotEmpty(t, data["updated_at"])
}

func TestRoomHandler_GetBySlug_ResponseHasAllFields(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Slug All", "color": "#222"}
	c, _ := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(5))
	require.NoError(t, h.Create(c))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/slug/slug-all", nil, map[string]string{"slug": "slug-all"})
	require.NoError(t, h.GetBySlug(c2))

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})

	assert.NotZero(t, data["id"])
	assert.Equal(t, float64(5), data["owner_id"])
	assert.Equal(t, "Slug All", data["name"])
	assert.Equal(t, "slug-all", data["slug"])
	assert.NotEmpty(t, data["created_at"])
}

func TestRoomHandler_List_ResponseItemsHaveAllFields(t *testing.T) {
	h, _ := setupRoomHandler(t)

	for i := 0; i < 2; i++ {
		body := map[string]string{"name": "List Room " + strconv.Itoa(i), "color": "#FFF"}
		c, _ := newEchoContext(http.MethodPost, "/api/rooms", body, nil)
		c.Set("user_id", int64(1))
		require.NoError(t, h.Create(c))
	}

	c, rec := newEchoContext(http.MethodGet, "/api/rooms?page=1&per_page=10", nil, nil)
	require.NoError(t, h.List(c))

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	require.Len(t, items, 2)

	for _, item := range items {
		room := item.(map[string]interface{})
		assert.NotZero(t, room["id"])
		assert.NotZero(t, room["owner_id"], "owner_id must be in list items")
		assert.NotEmpty(t, room["name"])
		assert.NotEmpty(t, room["slug"])
		assert.NotEmpty(t, room["created_at"])
		assert.NotEmpty(t, room["updated_at"])
	}
}

func TestRoomHandler_GetUsers_ResponseHasUserData(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := usecase.NewRoomUseCase(roomRepo, userRepo, sessionRepo)
	h := NewRoomHandler(uc)

	createBody := map[string]string{"name": "Room Users"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	roomID := int64(createResp["data"].(map[string]interface{})["id"].(float64))

	// Create users in the users table first
	u1 := &entity.User{Email: "u20@test.com", DisplayName: "Teacher User", Role: "teacher"}
	u2 := &entity.User{Email: "u21@test.com", DisplayName: "Student User", Role: "student"}
	_ = userRepo.Create(u1)
	_ = userRepo.Create(u2)

	h.roomUC.AddUser(roomID, u1.ID, "teacher")
	h.roomUC.AddUser(roomID, u2.ID, "student")

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/users", nil,
		map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.GetUsers(c2))

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &resp))
	users := resp["data"].([]interface{})
	require.Len(t, users, 2)
}

func TestRoomHandler_GetSettings_ResponseHasAllFields(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Settings Room"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	roomID := int64(createResp["data"].(map[string]interface{})["id"].(float64))

	c2, rec2 := newEchoContext(http.MethodGet, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/settings", nil,
		map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.GetSettings(c2))

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &resp))
	settings := resp["data"].(map[string]interface{})

	assert.NotZero(t, settings["max_users"], "max_users must be present")
	assert.NotNil(t, settings["recording_enabled"], "recording_enabled must be present")
	assert.NotNil(t, settings["allow_student_video"], "allow_student_video must be present")
	assert.NotNil(t, settings["allow_student_audio"], "allow_student_audio must be present")
	assert.NotNil(t, settings["allow_student_screen_share"], "allow_student_screen_share must be present")
	assert.NotNil(t, settings["allow_student_whiteboard"], "allow_student_whiteboard must be present")
	assert.NotNil(t, settings["allow_student_chat"], "allow_student_chat must be present")
	assert.NotZero(t, settings["session_auto_end_minutes"], "session_auto_end_minutes must be present")
	assert.NotNil(t, settings["waiting_room_enabled"], "waiting_room_enabled must be present")
}

func TestRoomHandler_UpdateSettings_PreservesAllFields(t *testing.T) {
	h, _ := setupRoomHandler(t)

	createBody := map[string]string{"name": "Settings Room"}
	c, rec := newEchoContext(http.MethodPost, "/api/rooms", createBody, nil)
	c.Set("user_id", int64(1))
	require.NoError(t, h.Create(c))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	roomID := int64(createResp["data"].(map[string]interface{})["id"].(float64))

	settingsBody := map[string]interface{}{
		"max_users":                  150,
		"recording_enabled":          false,
		"allow_student_video":        true,
		"allow_student_audio":        false,
		"allow_student_screen_share": true,
		"allow_student_whiteboard":   true,
		"allow_student_chat":         false,
		"session_auto_end_minutes":   45,
		"waiting_room_enabled":       true,
	}
	c2, rec2 := newEchoContext(http.MethodPut, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/settings", settingsBody,
		map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.UpdateSettings(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Verify by reading back
	c3, rec3 := newEchoContext(http.MethodGet, "/api/rooms/"+strconv.FormatInt(roomID, 10)+"/settings", nil,
		map[string]string{"id": strconv.FormatInt(roomID, 10)})
	require.NoError(t, h.GetSettings(c3))

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec3.Body.Bytes(), &resp))
	settings := resp["data"].(map[string]interface{})

	assert.Equal(t, float64(150), settings["max_users"])
	assert.Equal(t, false, settings["recording_enabled"])
	assert.Equal(t, true, settings["allow_student_video"])
	assert.Equal(t, false, settings["allow_student_audio"])
	assert.Equal(t, true, settings["allow_student_screen_share"])
	assert.Equal(t, true, settings["allow_student_whiteboard"])
	assert.Equal(t, false, settings["allow_student_chat"])
	assert.Equal(t, float64(45), settings["session_auto_end_minutes"])
	assert.Equal(t, true, settings["waiting_room_enabled"])
}
