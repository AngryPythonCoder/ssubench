CREATE TYPE task_status as ENUM ('published', 'in_progress', 'done', 'completed', 'canceled');

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    reward INT NOT NULL CHECK (reward >= 0),
    status task_status DEFAULT 'published',
    customer_id INT NOT NULL REFERENCES users(id),
    performer_id INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bids (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    performer_id INT NOT NULL REFERENCES users(id),
    text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (task_id, performer_id)
);

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id),
    customer_id INT NOT NULL REFERENCES users(id),
    performer_id INT NOT NULL REFERENCES users(id),
    amount INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);