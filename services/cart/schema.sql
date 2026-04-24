-- Cart Items Table
CREATE TABLE IF NOT EXISTS cart_items (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  cart_key VARCHAR(255) NOT NULL COMMENT 'user_id (INT as str) or session_id',
  product_id BIGINT NOT NULL,
  variant JSON NULL COMMENT 'e.g. {"color": "red", "size": "M"}',
  quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_cart_key (cart_key),
  INDEX idx_product (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Seed example (optional)
-- INSERT INTO cart_items (cart_key, product_id, quantity) VALUES ('guest_123', 1, 2);

