package services

import (
	"fmt"
	"html"
	"log"
	"net/smtp"
	"strings"
	"sync"
)

// Email represents an email message to be sent.
type Email struct {
	To      string
	Subject string
	Body    string
}

// EmailService handles sending emails via SMTP with async queue-based delivery.
type EmailService struct {
	config  SMTPSettings
	queue   chan Email
	wg      sync.WaitGroup
	workers int
}

// SMTPSettings holds the SMTP configuration for the email service.
type SMTPSettings struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Enabled  bool
}

// NewEmailService creates a new email service with the given SMTP settings.
// It starts a background worker goroutine that processes the email queue.
func NewEmailService(settings SMTPSettings, workers int, queueSize int) *EmailService {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 1 {
		queueSize = 100
	}

	svc := &EmailService{
		config:  settings,
		queue:   make(chan Email, queueSize),
		workers: workers,
	}

	for i := 0; i < workers; i++ {
		svc.wg.Add(1)
		go svc.worker(i)
	}

	return svc
}

// worker processes emails from the queue.
func (s *EmailService) worker(id int) {
	defer s.wg.Done()
	for email := range s.queue {
		s.send(email)
	}
}

// send delivers a single email. When SMTP is disabled, it logs the email instead.
func (s *EmailService) send(email Email) {
	if !s.config.Enabled {
		log.Printf("[EMAIL DISABLED] To: %s | Subject: %s", email.To, email.Subject)
		return
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	headers := make(map[string]string)
	headers["From"] = s.config.From
	headers["To"] = email.To
	headers["Subject"] = email.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(email.Body)

	var auth smtp.Auth
	if s.config.Username != "" && s.config.Password != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	err := smtp.SendMail(addr, auth, s.config.From, []string{email.To}, []byte(msg.String()))
	if err != nil {
		log.Printf("[EMAIL ERROR] Failed to send email to %s: %v", email.To, err)
	} else {
		log.Printf("[EMAIL SENT] To: %s | Subject: %s", email.To, email.Subject)
	}
}

// Enqueue adds an email to the sending queue.
func (s *EmailService) Enqueue(email Email) {
	select {
	case s.queue <- email:
	default:
		log.Printf("[EMAIL QUEUE FULL] Dropping email to %s", email.To)
	}
}

// Shutdown gracefully shuts down the email service, waiting for queued emails to be sent.
func (s *EmailService) Shutdown() {
	close(s.queue)
	s.wg.Wait()
}

// --- Template helpers ---

// SendWelcome sends a welcome email to a new user.
func (s *EmailService) SendWelcome(to, fullName string) {
	subject := "خوش آمدید به آی‌روم"
	body := welcomeTemplate(fullName)
	s.Enqueue(Email{To: to, Subject: subject, Body: body})
}

// SendPasswordReset sends a password reset email with a reset link.
func (s *EmailService) SendPasswordReset(to, fullName, resetLink string) {
	subject := "بازنشانی رمز عبور - آی‌روم"
	body := passwordResetTemplate(fullName, resetLink)
	s.Enqueue(Email{To: to, Subject: subject, Body: body})
}

// SendSessionReminder sends a session reminder email (15 minutes before start).
func (s *EmailService) SendSessionReminder(to, fullName, sessionTitle, sessionTime, joinLink string) {
	subject := "یادآوری جلسه - آی‌روم"
	body := sessionReminderTemplate(fullName, sessionTitle, sessionTime, joinLink)
	s.Enqueue(Email{To: to, Subject: subject, Body: body})
}

// SendTicketReply sends a ticket reply notification email.
func (s *EmailService) SendTicketReply(to, fullName, ticketTitle, replyPreview, ticketLink string) {
	subject := "پاسخ جدید به تیکت شما - آی‌روم"
	body := ticketReplyTemplate(fullName, ticketTitle, replyPreview, ticketLink)
	s.Enqueue(Email{To: to, Subject: subject, Body: body})
}

// SendClassEnrollment sends a class enrollment notification email.
func (s *EmailService) SendClassEnrollment(to, fullName, className, teacherName, classLink string) {
	subject := "ثبت‌نام در کلاس جدید - آی‌روم"
	body := classEnrollmentTemplate(fullName, className, teacherName, classLink)
	s.Enqueue(Email{To: to, Subject: subject, Body: body})
}

// SendAnnouncement sends an announcement notification email.
func (s *EmailService) SendAnnouncement(to, fullName, title, content, link string) {
	subject := "اعلان جدید - آی‌روم"
	body := announcementTemplate(fullName, title, content, link)
	s.Enqueue(Email{To: to, Subject: subject, Body: body})
}

func escapeHTML(s string) string {
	return html.EscapeString(s)
}

// --- HTML Templates (Persian) ---

func welcomeTemplate(fullName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="fa">
<head><meta charset="UTF-8"></head>
<body style="font-family: Tahoma, Arial, sans-serif; background: #f5f5f5; padding: 20px;">
<div style="max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
<div style="background: #4F46E5; padding: 30px; text-align: center;">
<h1 style="color: #fff; margin: 0; font-size: 24px;">به آی‌روم خوش آمدید</h1>
</div>
<div style="padding: 30px;">
<p style="font-size: 16px; color: #333;">سلام <strong>%s</strong>،</p>
<p style="font-size: 14px; color: #555; line-height: 1.8;">از عضویت شما در پلتفرم آی‌روم سپاسگزاریم. هم اکنون می‌توانید از امکانات کلاس آنلاین، جلسات و ابزارهای آموزشی استفاده کنید.</p>
<div style="text-align: center; margin: 30px 0;">
<a href="#" style="background: #4F46E5; color: #fff; padding: 12px 30px; border-radius: 6px; text-decoration: none; font-size: 14px;">ورود به پنل کاربری</a>
</div>
<p style="font-size: 12px; color: #999;">در صورت داشتن هرگونه سوال، با پشتیبانی ما تماس بگیرید.</p>
</div>
</div>
</body>
</html>`, escapeHTML(fullName))
}

func passwordResetTemplate(fullName, resetLink string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="fa">
<head><meta charset="UTF-8"></head>
<body style="font-family: Tahoma, Arial, sans-serif; background: #f5f5f5; padding: 20px;">
<div style="max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
<div style="background: #DC2626; padding: 30px; text-align: center;">
<h1 style="color: #fff; margin: 0; font-size: 24px;">بازنشانی رمز عبور</h1>
</div>
<div style="padding: 30px;">
<p style="font-size: 16px; color: #333;">سلام <strong>%s</strong>،</p>
<p style="font-size: 14px; color: #555; line-height: 1.8;">درخواست بازنشانی رمز عبور برای حساب شما دریافت شده است. برای تنظیم رمز عبور جدید، روی دکمه زیر کلیک کنید:</p>
<div style="text-align: center; margin: 30px 0;">
<a href="%s" style="background: #DC2626; color: #fff; padding: 12px 30px; border-radius: 6px; text-decoration: none; font-size: 14px;">بازنشانی رمز عبور</a>
</div>
<p style="font-size: 12px; color: #999;">این لینک به مدت ۱ ساعت معتبر است. اگر شما این درخواست را ارسال نکرده‌اید، این ایمیل را نادیده بگیرید.</p>
</div>
</div>
</body>
</html>`, escapeHTML(fullName), resetLink)
}

func sessionReminderTemplate(fullName, sessionTitle, sessionTime, joinLink string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="fa">
<head><meta charset="UTF-8"></head>
<body style="font-family: Tahoma, Arial, sans-serif; background: #f5f5f5; padding: 20px;">
<div style="max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
<div style="background: #F59E0B; padding: 30px; text-align: center;">
<h1 style="color: #fff; margin: 0; font-size: 24px;">یادآوری جلسه</h1>
</div>
<div style="padding: 30px;">
<p style="font-size: 16px; color: #333;">سلام <strong>%s</strong>،</p>
<p style="font-size: 14px; color: #555; line-height: 1.8;">جلسه زیر تا <strong>۱۵ دقیقه</strong> دیگر شروع می‌شود:</p>
<div style="background: #FFFBEB; border: 1px solid #FCD34D; border-radius: 6px; padding: 15px; margin: 20px 0;">
<p style="margin: 0; font-size: 16px; color: #92400E;"><strong>عنوان:</strong> %s</p>
<p style="margin: 8px 0 0; font-size: 14px; color: #92400E;"><strong>زمان:</strong> %s</p>
</div>
<div style="text-align: center; margin: 30px 0;">
<a href="%s" style="background: #F59E0B; color: #fff; padding: 12px 30px; border-radius: 6px; text-decoration: none; font-size: 14px;">ورود به جلسه</a>
</div>
</div>
</div>
</body>
</html>`, escapeHTML(fullName), escapeHTML(sessionTitle), escapeHTML(sessionTime), joinLink)
}

func ticketReplyTemplate(fullName, ticketTitle, replyPreview, ticketLink string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="fa">
<head><meta charset="UTF-8"></head>
<body style="font-family: Tahoma, Arial, sans-serif; background: #f5f5f5; padding: 20px;">
<div style="max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
<div style="background: #059669; padding: 30px; text-align: center;">
<h1 style="color: #fff; margin: 0; font-size: 24px;">پاسخ جدید به تیکت</h1>
</div>
<div style="padding: 30px;">
<p style="font-size: 16px; color: #333;">سلام <strong>%s</strong>،</p>
<p style="font-size: 14px; color: #555; line-height: 1.8;">پاسخ جدیدی به تیکت شما ارسال شده است:</p>
<div style="background: #ECFDF5; border: 1px solid #6EE7B7; border-radius: 6px; padding: 15px; margin: 20px 0;">
<p style="margin: 0; font-size: 14px; color: #065F46;"><strong>عنوان تیکت:</strong> %s</p>
<p style="margin: 8px 0 0; font-size: 13px; color: #047857; line-height: 1.6;">%s</p>
</div>
<div style="text-align: center; margin: 30px 0;">
<a href="%s" style="background: #059669; color: #fff; padding: 12px 30px; border-radius: 6px; text-decoration: none; font-size: 14px;">مشاهده تیکت</a>
</div>
</div>
</div>
</body>
</html>`, escapeHTML(fullName), escapeHTML(ticketTitle), escapeHTML(replyPreview), ticketLink)
}

func classEnrollmentTemplate(fullName, className, teacherName, classLink string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="fa">
<head><meta charset="UTF-8"></head>
<body style="font-family: Tahoma, Arial, sans-serif; background: #f5f5f5; padding: 20px;">
<div style="max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
<div style="background: #7C3AED; padding: 30px; text-align: center;">
<h1 style="color: #fff; margin: 0; font-size: 24px;">ثبت‌نام در کلاس جدید</h1>
</div>
<div style="padding: 30px;">
<p style="font-size: 16px; color: #333;">سلام <strong>%s</strong>،</p>
<p style="font-size: 14px; color: #555; line-height: 1.8;">شما در کلاس جدیدی ثبت‌نام شده‌اید:</p>
<div style="background: #F5F3FF; border: 1px solid #C4B5FD; border-radius: 6px; padding: 15px; margin: 20px 0;">
<p style="margin: 0; font-size: 16px; color: #5B21B6;"><strong>نام کلاس:</strong> %s</p>
<p style="margin: 8px 0 0; font-size: 14px; color: #6D28D9;"><strong>استاد:</strong> %s</p>
</div>
<div style="text-align: center; margin: 30px 0;">
<a href="%s" style="background: #7C3AED; color: #fff; padding: 12px 30px; border-radius: 6px; text-decoration: none; font-size: 14px;">مشاهده کلاس</a>
</div>
</div>
</div>
</body>
</html>`, escapeHTML(fullName), escapeHTML(className), escapeHTML(teacherName), classLink)
}

func announcementTemplate(fullName, title, content, link string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="fa">
<head><meta charset="UTF-8"></head>
<body style="font-family: Tahoma, Arial, sans-serif; background: #f5f5f5; padding: 20px;">
<div style="max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
<div style="background: #2563EB; padding: 30px; text-align: center;">
<h1 style="color: #fff; margin: 0; font-size: 24px;">اعلان جدید</h1>
</div>
<div style="padding: 30px;">
<p style="font-size: 16px; color: #333;">سلام <strong>%s</strong>،</p>
<div style="background: #EFF6FF; border: 1px solid #93C5FD; border-radius: 6px; padding: 15px; margin: 20px 0;">
<p style="margin: 0; font-size: 16px; color: #1E40AF;"><strong>%s</strong></p>
<p style="margin: 10px 0 0; font-size: 14px; color: #1E3A8A; line-height: 1.8;">%s</p>
</div>
<div style="text-align: center; margin: 30px 0;">
<a href="%s" style="background: #2563EB; color: #fff; padding: 12px 30px; border-radius: 6px; text-decoration: none; font-size: 14px;">مشاهده جزئیات</a>
</div>
</div>
</div>
</body>
</html>`, escapeHTML(fullName), escapeHTML(title), escapeHTML(content), link)
}
