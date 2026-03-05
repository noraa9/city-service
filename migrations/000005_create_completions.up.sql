CREATE TABLE completions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id UUID REFERENCES requests(id) ON DELETE CASCADE,
  days_spent INT NOT NULL CHECK (days_spent > 0),
  comment    TEXT,                         -- contractor's comment about monitor's work
  photo_url  TEXT NOT NULL,                -- photo of completed work (MinIO URL)
  created_at TIMESTAMP DEFAULT NOW()
);

