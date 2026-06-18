package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/iroom/iroom/internal/domain/entity"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

const (
	webhookTimeout      = 10 * time.Second
	webhookMaxRetries   = 3
	webhookRetryDelay   = 2 * time.Second
)

type WebhookDeliveryService struct {
	webhookRepo    *repository.WebhookRepo
	deliveryRepo   *repository.WebhookDeliveryRepo
	client         *http.Client
}

func NewWebhookDeliveryService(webhookRepo *repository.WebhookRepo, deliveryRepo *repository.WebhookDeliveryRepo) *WebhookDeliveryService {
	return &WebhookDeliveryService{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		client: &http.Client{
			Timeout: webhookTimeout,
		},
	}
}

// DispatchEvent sends the event to all webhooks subscribed to this event type
func (s *WebhookDeliveryService) DispatchEvent(eventType string, data interface{}) {
	webhooks, err := s.webhookRepo.GetActiveByEventType(eventType)
	if err != nil {
		slog.Error("failed to get webhooks for event", "event", eventType, "error", err)
		return
	}

	event := entity.WebhookEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	for _, webhook := range webhooks {
		go s.deliverWebhook(webhook, event)
	}
}

// deliverWebhook sends a webhook payload with retry logic
func (s *WebhookDeliveryService) deliverWebhook(webhook entity.Webhook, event entity.WebhookEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal webhook payload", "error", err)
		return
	}

	var lastStatusCode *int
	var lastResponseBody string
	var success bool

	for attempt := 0; attempt <= webhookMaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2s, 4s, 8s
			backoff := webhookRetryDelay * time.Duration(1<<uint(attempt-1))
			slog.Info("retrying webhook delivery", "webhook_id", webhook.ID, "attempt", attempt, "backoff", backoff)
			time.Sleep(backoff)
		}

		statusCode, responseBody, err := s.sendWebhook(webhook, payload)
		if err != nil {
			slog.Error("webhook delivery failed", "webhook_id", webhook.ID, "attempt", attempt, "error", err)
			lastStatusCode = nil
			lastResponseBody = err.Error()
			success = false
			continue
		}

		lastStatusCode = &statusCode
		lastResponseBody = responseBody
		success = statusCode >= 200 && statusCode < 300

		if success {
			slog.Info("webhook delivered successfully", "webhook_id", webhook.ID, "status", statusCode)
			break
		}

		slog.Warn("webhook delivery returned non-success status", "webhook_id", webhook.ID, "status", statusCode, "attempt", attempt)
	}

	// Log the delivery
	delivery := &entity.WebhookDelivery{
		WebhookID:    webhook.ID,
		EventType:    event.Type,
		Payload:      string(payload),
		StatusCode:   *lastStatusCode,
		ResponseBody: lastResponseBody,
		Success:      success,
		RetryCount:   0,
	}

	if !success {
		delivery.RetryCount = webhookMaxRetries
	}

	if err := s.deliveryRepo.Create(delivery); err != nil {
		slog.Error("failed to log webhook delivery", "error", err)
	}
}

// sendWebhook sends a single webhook request
func (s *WebhookDeliveryService) sendWebhook(webhook entity.Webhook, payload []byte) (int, string, error) {
	req, err := http.NewRequest(http.MethodPost, webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-ID", fmt.Sprintf("%d", webhook.ID))
	req.Header.Set("X-Webhook-Signature", generateSignature(payload, webhook.Secret))
	req.Header.Set("User-Agent", "IRoom-Webhook/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit to 1MB
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, string(body), nil
}

// generateSignature creates HMAC-SHA256 signature for webhook payload
func generateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// SendTestEvent sends a test event to a specific webhook
func (s *WebhookDeliveryService) SendTestEvent(webhook entity.Webhook) {
	event := entity.WebhookEvent{
		Type:      "test",
		Timestamp: time.Now(),
		Data: map[string]string{
			"message": "This is a test event from IRoom",
		},
	}

	go s.deliverWebhook(webhook, event)
}
