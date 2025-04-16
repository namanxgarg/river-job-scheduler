CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  payload TEXT,
  status TEXT,
  result TEXT,
  retries INT DEFAULT 0,
  max_retries INT DEFAULT 3,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
