CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS states (
  id SMALLSERIAL PRIMARY KEY,
  code VARCHAR(2) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_states_code_upper CHECK (code = UPPER(code))
);

CREATE TABLE IF NOT EXISTS cities (
  id BIGSERIAL PRIMARY KEY,
  state_id SMALLINT NOT NULL REFERENCES states(id),
  name TEXT NOT NULL,
  name_normalized TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(state_id, name_normalized)
);

CREATE TABLE IF NOT EXISTS neighborhoods (
  id BIGSERIAL PRIMARY KEY,
  city_id BIGINT NOT NULL REFERENCES cities(id),
  name TEXT NOT NULL,
  name_normalized TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(city_id, name_normalized)
);

CREATE TABLE IF NOT EXISTS pharmacies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cnpj VARCHAR(14) NOT NULL UNIQUE,
  name TEXT NOT NULL,
  address TEXT NOT NULL,
  neighborhood_id BIGINT NOT NULL REFERENCES neighborhoods(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_pharmacies_cnpj_digits CHECK (cnpj ~ '^[0-9]{14}$')
);

CREATE INDEX IF NOT EXISTS idx_cities_state ON cities(state_id);
CREATE INDEX IF NOT EXISTS idx_cities_name_norm ON cities(name_normalized);
CREATE INDEX IF NOT EXISTS idx_neighborhoods_city ON neighborhoods(city_id);
CREATE INDEX IF NOT EXISTS idx_neighborhoods_name_norm ON neighborhoods(name_normalized);
CREATE INDEX IF NOT EXISTS idx_pharmacies_neighborhood ON pharmacies(neighborhood_id);
