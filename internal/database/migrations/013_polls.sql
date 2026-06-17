CREATE TABLE IF NOT EXISTS polls (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  question TEXT NOT NULL,
  options TEXT NOT NULL, -- JSON array of strings
  is_active BOOLEAN DEFAULT TRUE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_polls_session ON polls(session_id);

CREATE TABLE IF NOT EXISTS poll_votes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  poll_id INTEGER NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES users(id),
  option_index INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(poll_id, user_id)
);
CREATE INDEX idx_poll_votes_poll ON poll_votes(poll_id);
