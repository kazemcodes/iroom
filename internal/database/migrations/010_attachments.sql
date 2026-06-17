CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id INTEGER REFERENCES messages(id) ON DELETE CASCADE,
    ticket_id INTEGER REFERENCES tickets(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attachments_message ON attachments(message_id);
CREATE INDEX IF NOT EXISTS idx_attachments_ticket ON attachments(ticket_id);
