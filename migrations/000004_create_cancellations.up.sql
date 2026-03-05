CREATE TABLE cancellations (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id UUID REFERENCES requests(id) ON DELETE CASCADE,
  reason     VARCHAR(100) NOT NULL
             CHECK (reason IN ('not_relevant', 'wrong_data', 'mistake', 'other')),
  comment    TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

