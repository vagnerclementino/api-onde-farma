CREATE TEMP TABLE tmp_pharmacies (
  cnpj TEXT,
  name TEXT,
  address TEXT,
  neighborhood TEXT
);

\copy tmp_pharmacies(cnpj, name, address, neighborhood) FROM 'src/data/pharmacies.csv' WITH (FORMAT csv, HEADER true)

INSERT INTO states(code)
VALUES ('MG')
ON CONFLICT (code) DO NOTHING;

INSERT INTO cities(state_id, name, name_normalized)
SELECT s.id, 'BELO HORIZONTE', 'BELO HORIZONTE'
FROM states s
WHERE s.code = 'MG'
ON CONFLICT (state_id, name_normalized) DO NOTHING;

INSERT INTO neighborhoods(city_id, name, name_normalized)
SELECT c.id, TRIM(t.neighborhood), UPPER(TRIM(t.neighborhood))
FROM tmp_pharmacies t
JOIN cities c ON c.name_normalized = 'BELO HORIZONTE'
GROUP BY c.id, TRIM(t.neighborhood), UPPER(TRIM(t.neighborhood))
ON CONFLICT (city_id, name_normalized) DO NOTHING;

INSERT INTO pharmacies(cnpj, name, address, neighborhood_id)
SELECT
  REGEXP_REPLACE(t.cnpj, '[^0-9]', '', 'g') AS cnpj,
  TRIM(t.name) AS name,
  TRIM(t.address) AS address,
  n.id AS neighborhood_id
FROM tmp_pharmacies t
JOIN neighborhoods n ON n.name_normalized = UPPER(TRIM(t.neighborhood))
JOIN cities c ON c.id = n.city_id AND c.name_normalized = 'BELO HORIZONTE'
ON CONFLICT (cnpj) DO UPDATE
SET
  name = EXCLUDED.name,
  address = EXCLUDED.address,
  neighborhood_id = EXCLUDED.neighborhood_id,
  updated_at = NOW();

REFRESH MATERIALIZED VIEW pharmacy_search_mv;
