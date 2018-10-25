DROP DATABASE IF EXISTS db1;
CREATE DATABASE db1;

DROP DATABASE IF EXISTS db2;
CREATE DATABASE db2;

DROP DATABASE IF EXISTS db3;
CREATE DATABASE db3;

\c db1

CREATE TABLE table1(
    id integer NOT NULL,
    name varchar(255) NOT NULL
);

INSERT INTO table1 (id, name)
VALUES
(1, 'John'),
(2, 'George'),
(3, 'Richard');

\c db2

CREATE TABLE table1(
    id integer NOT NULL,
    name varchar(255) NOT NULL
);

INSERT INTO table1 (id, name)
VALUES
(1, 'Rob'),
(2, 'Ken'),
(3, 'Robert');

\c db3

CREATE TABLE table1(
    id integer NOT NULL,
    name varchar(255) NOT NULL
);

INSERT INTO table1 (id, name)
VALUES
(1, 'Athos'),
(2, 'Porthos'),
(3, 'Aramis');
