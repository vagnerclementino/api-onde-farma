CREATE MATERIALIZED VIEW IF NOT EXISTS pharmacy_search_mv AS
SELECT
  p.id::text AS id,
  p.cnpj,
  p.name,
  p.address,
  n.name AS neighborhood,
  c.name AS city,
  s.code AS state,
  n.name_normalized AS neighborhood_normalized,
  c.name_normalized AS city_normalized,
  s.code AS state_normalized
FROM pharmacies p
JOIN neighborhoods n ON n.id = p.neighborhood_id
JOIN cities c ON c.id = n.city_id
JOIN states s ON s.id = c.state_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_pharmacy_search_mv_id ON pharmacy_search_mv(id);
CREATE INDEX IF NOT EXISTS idx_pharmacy_search_mv_cnpj ON pharmacy_search_mv(cnpj);
CREATE INDEX IF NOT EXISTS idx_pharmacy_search_mv_state_city_neighborhood
  ON pharmacy_search_mv(state_normalized, city_normalized, neighborhood_normalized);
