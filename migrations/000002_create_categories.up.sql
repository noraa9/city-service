CREATE TABLE categories (
  id   SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(100) UNIQUE NOT NULL
);

INSERT INTO categories (name, slug) VALUES
  ('Нарушение благоустройства', 'blagoustrojstvo'),
  ('Твердо Бытовые Отходы',     'tbo'),
  ('Светофоры и ТСРДД',         'svetofory'),
  ('Очистка снега',             'sneg'),
  ('Другое',                    'other');

