CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name          VARCHAR(255) NOT NULL,
  email              VARCHAR(255) UNIQUE NOT NULL,
  password_hash      TEXT NOT NULL,
  phone              VARCHAR(20),
  role               VARCHAR(20) NOT NULL CHECK (role IN ('monitor', 'contractor', 'admin')),
  -- contractor-only fields:
  company_name       VARCHAR(255),        -- e.g. "ИП CleanUralskCar"
  responsible_person VARCHAR(255),        -- e.g. "Копберген Ержан"
  company_phone      VARCHAR(20),         -- responsible person's phone
  created_at         TIMESTAMP DEFAULT NOW()
);

