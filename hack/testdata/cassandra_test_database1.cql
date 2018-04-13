CREATE KEYSPACE space1 WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 2 };
USE space1;
CREATE TABLE testtable1 (key varchar, value varchar, PRIMARY KEY(key));
