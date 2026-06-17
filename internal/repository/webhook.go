package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/iroom/iroom/internal/models"
)

type WebhookRepo struct {
	db *sql.DB
}

func NewWebhookRepo(db *sql.DB) *WebhookRepo {
	return &WebhookRepo{db: db}
}

func (r *WebhookRepo) Create(w *models.Webhook) error {
	eventsJSON, err := json.Marshal(w.Events)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(
		`INSERT INTO webhooks (user_id, url, secret, events, is_active) VALUES (?, ?, ?, ?, ?)`,
		w.UserID, w.URL, w.Secret, string(eventsJSON), w.IsActive,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	w.ID = id
	return nil
}

func (r *WebhookRepo) GetByID(id int64) (*models.Webhook, error) {
	w := &models.Webhook{}
	var eventsJSON string
	err := r.db.QueryRow(
		`SELECT id, user_id, url, secret, events, is_active, created_at FROM webhooks WHERE id = ?`, id,
	).Scan(&w.ID, &w.UserID, &w.URL, &w.Secret, &eventsJSON, &w.IsActive, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(eventsJSON), &w.Events); err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WebhookRepo) ListByUserID(userID int64) ([]models.Webhook, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, url, secret, events, is_active, created_at FROM webhooks WHERE user_id = ? ORDER BY id DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		var eventsJSON string
		if err := rows.Scan(&w.ID, &w.UserID, &w.URL, &w.Secret, &eventsJSON, &w.IsActive, &w.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(eventsJSON), &w.Events); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (r *WebhookRepo) ListAll() ([]models.Webhook, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, url, secret, events, is_active, created_at FROM webhooks ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		var eventsJSON string
		if err := rows.Scan(&w.ID, &w.UserID, &w.URL, &w.Secret, &eventsJSON, &w.IsActive, &w.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(eventsJSON), &w.Events); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (r *WebhookRepo) GetActiveByEventType(eventType string) ([]models.Webhook, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, url, secret, events, is_active, created_at FROM webhooks WHERE is_active = TRUE`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		var eventsJSON string
		if err := rows.Scan(&w.ID, &w.UserID, &w.URL, &w.Secret, &eventsJSON, &w.IsActive, &w.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(eventsJSON), &w.Events); err != nil {
			return nil, err
		}
		// Check if this webhook subscribes to the event type
		for _, event := range w.Events {
			if event == eventType {
				webhooks = append(webhooks, w)
				break
			}
		}
	}
	return webhooks, nil
}

func (r *WebhookRepo) Update(w *models.Webhook) error {
	eventsJSON, err := json.Marshal(w.Events)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`UPDATE webhooks SET url = ?, events = ?, is_active = ? WHERE id = ?`,
		w.URL, string(eventsJSON), w.IsActive, w.ID,
	)
	return err
}

func (r *WebhookRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM webhooks WHERE id = ?`, id)
	return err
}

// Webhook Delivery methods

type WebhookDeliveryRepo struct {
	db *sql.DB
}

func NewWebhookDeliveryRepo(db *sql.DB) *WebhookDeliveryRepo {
	return &WebhookDeliveryRepo{db: db}
}

func (r *WebhookDeliveryRepo) Create(d *models.WebhookDelivery) error {
	result, err := r.db.Exec(
		`INSERT INTO webhook_deliveries (webhook_id, event_type, payload, status_code, response_body, success, retry_count) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.WebhookID, d.EventType, d.Payload, d.StatusCode, d.ResponseBody, d.Success, d.RetryCount,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	d.ID = id
	return nil
}

func (r *WebhookDeliveryRepo) ListByWebhookID(webhookID int64, page, perPage int) ([]models.WebhookDelivery, int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = ?`, webhookID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(
		`SELECT id, webhook_id, event_type, payload, status_code, response_body, success, retry_count, created_at 
		FROM webhook_deliveries WHERE webhook_id = ? ORDER BY id DESC LIMIT ? OFFSET ?`,
		webhookID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deliveries []models.WebhookDelivery
	for rows.Next() {
		var d models.WebhookDelivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.EventType, &d.Payload, &d.StatusCode, &d.ResponseBody, &d.Success, &d.RetryCount, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, total, nil
}
