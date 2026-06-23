CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin',
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cars (
    id SERIAL PRIMARY KEY,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    brand TEXT NOT NULL,
    model TEXT NOT NULL,
    year INTEGER,
    vin TEXT,
    license_plate TEXT,
    mileage INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (id, client_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    car_id INTEGER NOT NULL REFERENCES cars(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'new',
    problem_description TEXT NOT NULL,
    manager_comment TEXT,
    total_price NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    completed_at TIMESTAMP,
    CONSTRAINT orders_car_client_match
        FOREIGN KEY (car_id, client_id) REFERENCES cars(id, client_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_type TEXT NOT NULL CHECK (item_type IN ('service', 'part')),
    title TEXT NOT NULL,
    quantity NUMERIC(10,2) NOT NULL DEFAULT 1,
    price NUMERIC(12,2) NOT NULL DEFAULT 0,
    total NUMERIC(12,2) GENERATED ALWAYS AS (quantity * price) STORED
);

CREATE TABLE IF NOT EXISTS service_settings (
    id SERIAL PRIMARY KEY,
    company_name TEXT NOT NULL DEFAULT 'AutoService CRM',
    address TEXT NOT NULL DEFAULT 'Service address',
    phone TEXT NOT NULL DEFAULT '+49 000 000000',
    email TEXT NOT NULL DEFAULT 'service@example.com'
);

INSERT INTO service_settings (company_name, address, phone, email)
SELECT 'AutoService CRM', 'Musterstraße 1, Berlin', '+49 000 000000', 'service@example.com'
WHERE NOT EXISTS (SELECT 1 FROM service_settings);
