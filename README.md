verify data no psql from docker container always :

sudo docker start volurya
sudo docker start volurya_postgres
sudo docker ps
sudo docker exec -it volurya_postgres psql -U volurya -d volurya_db

SELECT * FROM product;




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
