CREATE TABLE IF NOT EXISTS recurring_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    day_of_week INTEGER NOT NULL CHECK(day_of_week >= 0 AND day_of_week <= 6),
    start_time TEXT NOT NULL,
    duration INTEGER NOT NULL DEFAULT 60,
    week_count INTEGER NOT NULL DEFAULT 12,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
