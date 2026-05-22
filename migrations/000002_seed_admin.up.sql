INSERT INTO users (username, password_hash, role, status, balance)
VALUES ('admin', '$2a$10$5cZEcMvGky7GXp6eUuFT2.JlEYIqnVTR9kW4Axn/hfN3a7ESOZ.Ea', 'admin', 'active', 0)
ON CONFLICT (username) DO NOTHING;