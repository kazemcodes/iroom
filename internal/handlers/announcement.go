package handlers

import (
	"database/sql"
	"strconv"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/labstack/echo/v4"
)

type AnnouncementHandler struct {
	announcementRepo *repository.AnnouncementRepo
	classRepo        *repository.ClassRepo
	logRepo          *repository.ActivityLogRepo
}

func NewAnnouncementHandler(announcementRepo *repository.AnnouncementRepo, classRepo *repository.ClassRepo, logRepo *repository.ActivityLogRepo) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementRepo: announcementRepo,
		classRepo:        classRepo,
		logRepo:          logRepo,
	}
}

func (h *AnnouncementHandler) log(userID int64, action, entityType string, entityID int64, details, ip string) {
	h.logRepo.Create(&models.ActivityLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		IPAddress:  ip,
	})
}

func (h *AnnouncementHandler) Create(c echo.Context) error {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کلاس نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	userRole := getUserRole(c)

	// Check if user is teacher of this class or admin
	class, err := h.classRepo.GetByID(classID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "کلاس یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت کلاس")
	}

	if class.TeacherID != userID && userRole != "admin" {
		return response.Forbidden(c, "شما اجازه ایجاد اعلان در این کلاس را ندارید")
	}

	var req models.CreateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Title == "" || req.Content == "" {
		return response.BadRequest(c, "عنوان و محتوا الزامی هستند")
	}

	classIDPtr := classID
	announcement := &models.Announcement{
		ClassID:  &classIDPtr,
		AuthorID: userID,
		Title:    req.Title,
		Content:  req.Content,
		IsPinned: req.IsPinned,
	}

	if err := h.announcementRepo.Create(announcement); err != nil {
		return response.InternalError(c, "خطا در ایجاد اعلان")
	}

	h.log(userID, "create_announcement", "announcement", announcement.ID, req.Title, c.RealIP())

	return response.Created(c, announcement)
}

func (h *AnnouncementHandler) List(c echo.Context) error {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کلاس نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	if role != "admin" {
		class, err := h.classRepo.GetByID(classID)
		if err != nil {
			return response.NotFound(c, "کلاس یافت نشد")
		}
		if class.TeacherID != userID && !h.classRepo.IsEnrolled(class.ID, userID) {
			return response.Forbidden(c, "دسترسی غیرمجاز")
		}
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	announcements, total, err := h.announcementRepo.ListByClass(classID, page, perPage)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت اعلان‌ها")
	}
	if announcements == nil {
		announcements = []*models.Announcement{}
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      announcements,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *AnnouncementHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	userRole := getUserRole(c)

	announcement, err := h.announcementRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "اعلان یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت اعلان")
	}

	// Only author or admin can update
	if announcement.AuthorID != userID && userRole != "admin" {
		return response.Forbidden(c, "شما اجازه ویرایش این اعلان را ندارید")
	}

	var req models.UpdateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Title != "" {
		announcement.Title = req.Title
	}
	if req.Content != "" {
		announcement.Content = req.Content
	}

	if err := h.announcementRepo.Update(announcement); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی اعلان")
	}

	h.log(userID, "update_announcement", "announcement", announcement.ID, req.Title, c.RealIP())

	return response.Success(c, announcement)
}

func (h *AnnouncementHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	userRole := getUserRole(c)

	announcement, err := h.announcementRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "اعلان یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت اعلان")
	}

	// Only author or admin can delete
	if announcement.AuthorID != userID && userRole != "admin" {
		return response.Forbidden(c, "شما اجازه حذف این اعلان را ندارید")
	}

	if err := h.announcementRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف اعلان")
	}

	h.log(userID, "delete_announcement", "announcement", id, "", c.RealIP())

	return response.Success(c, map[string]string{"message": "اعلان حذف شد"})
}

func (h *AnnouncementHandler) Pin(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	userRole := getUserRole(c)

	announcement, err := h.announcementRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "اعلان یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت اعلان")
	}

	// Only teacher of the class or admin can pin
	if announcement.ClassID != nil {
		class, err := h.classRepo.GetByID(*announcement.ClassID)
		if err != nil {
			return response.InternalError(c, "خطا در دریافت کلاس")
		}
		if class.TeacherID != userID && userRole != "admin" {
			return response.Forbidden(c, "شما اجازه پین کردن این اعلان را ندارید")
		}
	} else if userRole != "admin" {
		// System-wide announcements can only be pinned by admin
		return response.Forbidden(c, "شما اجازه پین کردن این اعلان را ندارید")
	}

	// Toggle pin status
	newPinned := !announcement.IsPinned
	if err := h.announcementRepo.SetPinned(id, newPinned); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی اعلان")
	}

	action := "unpin_announcement"
	if newPinned {
		action = "pin_announcement"
	}
	h.log(userID, action, "announcement", id, "", c.RealIP())

	return response.Success(c, map[string]interface{}{
		"id":         id,
		"is_pinned":  newPinned,
		"message":    "وضعیت پین تغییر کرد",
	})
}
