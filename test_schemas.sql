DROP DATABASE IF EXISTS db1;
CREATE DATABASE db1;

DROP DATABASE IF EXISTS db2;
CREATE DATABASE db2;

DROP DATABASE IF EXISTS db3;
CREATE DATABASE db3;

CREATE TABLE db1.table1(
    id INT(11) NOT NULL,
    name VARCHAR(255) NOT NULL
) ENGINE InnoDB;

CREATE TABLE db2.table1(
    id INT(11) NOT NULL,
    name VARCHAR(255) NOT NULL
) ENGINE InnoDB;

CREATE TABLE db3.table1(
    id INT(11) NOT NULL,
    name VARCHAR(255) NOT NULL
) ENGINE InnoDB;

INSERT INTO db1.table1 (id, name) VALUES (1, "John"), (2, "George"), (3, "Richard");
INSERT INTO db2.table1 (id, name) VALUES (1, "Rob"), (2, "Ken"), (3, "Robert");
INSERT INTO db3.table1 (id, name) VALUES (1, "Athos"), (2, "Porthos"), (3, "Aramis");
