Volurya API

Volurya API is a personal learning laboratory created to simulate real-world backend practices — such as JWT authentication, Clean Architecture, Docker, automated tests, and CI/CD — applied to a simple e-commerce domain.

The project models a fictional store selling products related to the band Volurya (t-shirts, caps, socks, drumsticks, etc).

This is not a commercial product, but a focused backend showcase built for learning and architectural validation.

🎯 Project Goals

Practice backend concepts used in production systems

Apply a simplified Clean Architecture

Implement real JWT-based authentication

Work with a real PostgreSQL database (including in tests)

Understand the difference between unit and integration tests

Use Docker to standardize the environment

Build a project that is easy to explain and present as a showcase

🧱 Architecture Overview

The application follows a simplified Clean Architecture, organized into clear layers:

Controller
HTTP layer (Gin). Responsible for request/response handling only.

UseCase
Contains the business rules and application logic.
Critical invariants (e.g. passwords are never stored in plain text) live here.

Repository
Data access layer (PostgreSQL).

Models
Domain entities.

This separation keeps business logic isolated from frameworks and infrastructure details.

🐳 Running the Project with Docker
sudo docker compose up --build
sudo docker start volurya
sudo docker start volurya_postgres


Check running containers:

sudo docker ps

🗄️ Accessing the PostgreSQL Database

Enter the database container:

sudo docker exec -it volurya_postgres psql -U volurya -d volurya_db


Verify data:

SELECT * FROM product;
SELECT * FROM users;

🧾 Database Schema (Example)
Products table
CREATE TABLE IF NOT EXISTS product (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  description VARCHAR(200) NOT NULL,
  price NUMERIC(10,2) NOT NULL,
  stock INTEGER DEFAULT 0 NOT NULL
);

Sample data
INSERT INTO product (name, description, price, stock)
VALUES
('Camisa volurya', 'camiseta baby look m', 100.00, 25),
('Camisa volurya', 'camiseta machao m', 100.00, 40),
('Bone', 'aba aberta', 300.00, 10),
('Meias volurya', 'meias alien m', 200.00, 25);


Expected result:

 id |      name      |     description      | price  | stock
----+----------------+----------------------+--------+-------
  1 | Camisa volurya | camiseta baby look m | 100.00 |    25
  2 | Camisa volurya | camiseta machao m    | 100.00 |    40
  3 | Bone           | aba aberta           | 300.00 |    10
  4 | Meias volurya  | meias alien m        | 200.00 |    25

🔌 API Health Check
GET http://localhost:8000/ping


Expected response:

{ "message": "pong" }

🔐 JWT Authentication Flow
1️⃣ Signup (Create User)
POST http://localhost:8000/signup

{
  "email": "admin@volurya.com",
  "password": "123456"
}


Expected response:

{ "message": "user created" }


Verify in the database:

SELECT * FROM users;

2️⃣ Login (Generate JWT)
POST http://localhost:8000/login

{
  "email": "admin@volurya.com",
  "password": "123456"
}


Expected response:

{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}


Copy the token without quotes.

3️⃣ Protected Route (Without Token)
GET http://localhost:8000/products


Expected response:

{ "error": "unauthorized" }


If this does not fail, the JWT middleware is not active.

4️⃣ Protected Route (With Token)
GET http://localhost:8000/products


Headers:

Authorization: Bearer YOUR_TOKEN_HERE


Expected:

HTTP 200

List of products

5️⃣ Create Product (Protected Route)
POST http://localhost:8000/products


Headers:

Authorization: Bearer YOUR_TOKEN
Content-Type: application/json


Body:

{
  "name": "Volurya T-Shirt",
  "price": 99.90,
  "stock": 10
}

🧪 Automated Tests

Start the test database:

docker compose -f docker-compose.test.yml up -d


Run all tests:

go test ./...


Tests include:

Unit tests for business rules

Integration tests (HTTP + PostgreSQL)

Database setup and cleanup helpers

📌 Architectural Decisions

Simplified Clean Architecture

JWT for authentication

Real database in integration tests

No DTOs (avoid unnecessary complexity)

Monolithic design

Project intentionally frozen

❄️ Project Status

This project is intentionally frozen.
The current scope fully satisfies its learning and showcase goals.

Further changes would only make sense under a new context or requirements.

🧠 JWT Mental Model
POST /signup   → create user
POST /login    → generate JWT
JWT            → sent in Authorization header
Middleware     → validates token
UseCase        → business rules
Repository     → database

🏁 Final Note

This project represents:

conscious architectural decisions

real testing practices

a clear stopping point

It is meant to demonstrate how to think, not just how to code.