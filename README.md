“Volurya API é um laboratório pessoal para simular práticas de produção: auth JWT, clean architecture, docker, testes e CI/CD, aplicadas a um e-commerce simples.”



verify data no psql from docker container always :

sudo docker compose up --build
sudo docker start volurya
sudo docker start volurya_postgres

sudo docker ps
sudo docker exec -it volurya_postgres psql -U volurya -d volurya_db

SELECT * FROM product;

SELECT * FROM users;




sql para db :



create table IF NOT EXISTS product (
   id SERIAL primary key,
   name VARCHAR(50) not null,
   description VARCHAR(200) not null,
   price NUMERIC(10, 2) not null,
   
   stock INTEGER DEFAULT 0 not null  
);

insert into product(name,description, price, stock) values ('Camisa volurya','camiseta baby look m', 100.00, 25), ('Camisa volurya','camiseta machao m', 100.00, 40),('Bone','aba aberta', 300.00,10),('Meias volurya','meias alien m', 200.00, 25);


this should return :


volurya_db=# SELECT * FROM product;
 id |      name      |     description      | price  | stock 
----+----------------+----------------------+--------+-------
  1 | Camisa volurya | camiseta baby look m | 100.00 |    25
  2 | Camisa volurya | camiseta machao m    | 100.00 |    40
  3 | Bone           | aba aberta           | 300.00 |    10
  4 | Meias volurya  | meias alien m        | 200.00 |    25
(4 rows)



volurya_db=# SELECT * FROM users;
 id |       email       |                           password                           | role  |  created_at         
----+-------------------+--------------------------------------------------------------+-------+----------------------------
  1 | admin@volurya.com | $2a$10$jCPYl7hYbfP3LJRLGwyFs.Ko9iIBDiSMi3MeARQU..c.ljKdhci8C | admin | 2025-12-19 22:42:47.611219
(1 row)



comandos no postman para ver se as rotas crud funcionam :

//testar se a porta e docker estao ativos mesmo:
GET http://localhost:8000/ping

deve voltar isso:
{
  "message": "pong"
}


//testar crud
get localhost:8000/products //get all products
post localhost:8000/products // create a new product
get localhost:8000/products/1 // get one product
delete localhost:8000/products/3  //delete one
put localhost:8000/products/6  //update one

TESTE COMPLETO NO POSTMAN DO JWT — AUTH + PRODUCTS (NO POSTMAN)

Testar se a API está viva:

GET

http://localhost:8000/ping

 Esperado :
{
  "message": "pong"
}



Criar usuário (SIGNUP) ;

POST
http://localhost:8000/signup

Headers
Content-Type: application/json

Body (raw → JSON)
{
  "email": "admin@volurya.com",
  "password": "123456"
}

Esperado:
{
  "message": "user created"
}



👉 Confirme no banco:

SELECT * FROM users;


Se não apareceu → problema no repository.

Fazer login (LOGIN): 
POST
http://localhost:8000/login

Body
{
  "email": "admin@volurya.com",
  "password": "123456"
}

Esperado :
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}


Copia esse token inteiro (sem aspas).



Testar rota protegida SEM token (tem que falhar):
GET
http://localhost:8000/products

Esperado
{
  "error": "unauthorized"
}


Se NÃO falhar → middleware não está sendo usado.

Testar rota protegida COM token:
GET
http://localhost:8000/products

Headers
Authorization: Bearer SEU_TOKEN_AQUI


Exemplo:

Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

Esperado:

Lista de produtos (vazia ou não)

Status 200

Middleware + JWT funcionando

Criar produto (rota protegida):
POST
http://localhost:8000/products

Headers
Authorization: Bearer SEU_TOKEN
Content-Type: application/json

Body
{
  "name": "Camiseta Volurya",
  "price": 99.90,
  "stock": 10
}

Esperado

Produto criado (ou objeto retornado):
Debug rápido (se algo der errado)
🔸 Sempre olhe:

Terminal (logs do Gin)

Mensagem exata do erro

Se o token está exatamente como Bearer <token>

🔸 Token inválido geralmente é:

token com espaço a mais

esqueceu Bearer

token expirado

Mapa mental final do jwt
POST /signup   → cria usuário
POST /login    → gera JWT
JWT            → vai no header
Middleware     → valida token
Usecase        → regra
Repository     → banco



testes :

sudo docker compose up --build
sudo docker start volurya_postgres_test


# 1. Subir banco de teste
docker compose -f docker-compose.test.yml up -d

# 2. Subir API com hot reload
docker compose up --build

# 3. Testar endpoint
curl http://localhost:8000/ping

subir usuario teste

INSERT INTO users (email, password, role) VALUES ('admin@volurya.com ', '$2a$10$jCPYl7hYbfP3LJRLGwyFs.Ko9iIBDiSMi3MeARQU..c.ljKdhci8C', 'admin ');
