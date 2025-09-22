INSERT INTO products
  (id, name)
VALUES
  (1, 'ProductA'),
  (2, 'ProductB'),
  (3, 'ProductC')
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name;

INSERT INTO menus
  (id, name)
VALUES
  ('1', 'MenusA'),
  ('2', 'MenusB'),
  ('3', 'MenusC')
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name;
