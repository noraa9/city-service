CREATE TABLE requests (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_number VARCHAR(20) UNIQUE NOT NULL,
  -- Format: {seq}{role_letter}{YYMMDD}
  -- Example: "1U202512" where U=monitor, seq=1, date=2025-12-xx
  -- Role letters: U = monitor, S = contractor
  title          VARCHAR(255) NOT NULL,
  category_id    INT REFERENCES categories(id),
  description    TEXT NOT NULL,
  urgency        VARCHAR(20) CHECK (urgency IN ('low', 'medium', 'critical')),
  deadline       DATE,
  location       TEXT NOT NULL,
  photo_url      TEXT,                    -- photo of the problem uploaded at creation
  status         VARCHAR(20) DEFAULT 'new'
                 CHECK (status IN ('new', 'in_progress', 'done', 'cancelled')),
  user_id        UUID REFERENCES users(id),
  contractor_id  UUID REFERENCES users(id),
  taken_at       TIMESTAMP,               -- when contractor took the request
  created_at     TIMESTAMP DEFAULT NOW(),
  updated_at     TIMESTAMP DEFAULT NOW()
);

-- auto-update updated_at on every update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ language 'plpgsql';

CREATE TRIGGER update_requests_updated_at
BEFORE UPDATE ON requests
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

