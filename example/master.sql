INSERT INTO products
  (id, name)
VALUES
  (1, 'ProductA'),
  (2, 'ProductB'),
  (3, 'ProductC')
ON CONFLICT (id) DO UPDATE
  SET name = EXCLUDED.name;
