CREATE SEQUENCE employee_seq;

CREATE TABLE employee (
  id int check (id > 0) NOT NULL DEFAULT NEXTVAL ('employee_seq'),
  name varchar(30) NOT NULL,
  city varchar(30) NOT NULL,
  PRIMARY KEY (id)
)  ;

ALTER SEQUENCE employee_seq RESTART WITH 1;