# First we create a project namespace
kubectl create ns project-x

# Now we install navigator in that namespace
helm \
    --debug \
    install \
    --wait \
    --name navigator \
    --namespace project-x \
    --set apiserver.image.repository=gcr.io/jetstack-sandbox/navigator-apiserver \
    --set controller.image.repository=gcr.io/jetstack-sandbox/navigator-controller \
    contrib/charts/navigator

# Let's see what's been installed
kubectl get all --namespace project-x
kubectl get clusterrolebinding --namespace project-x | grep navigator

# And let's take a look at the Navigator logs
# First the apiserver logs
kubectl --namespace project-x logs -c apiserver deploy/navigator-navigator-apiserver

# And then the controller logs
kubectl --namespace project-x logs deploy/navigator-navigator-controller

# Now that Navigator is running we can start a Cassandra single node cluster, using Helm.
helm install \
     --debug \
     --wait \
     --name "cc-1" \
     --namespace "project-x" \
     --set pilotImage.repository=gcr.io/jetstack-sandbox/navigator-pilot-cassandra \
     --set replicaCount=1 \
     contrib/charts/cassandra

# Let's take a quick look at the CassandraCluster resource
kubectl --namespace project-x get cassandraclusters --output yaml

# And let's see what other resources have been created
kubectl --namespace project-x get all

# We can look at the cassandra database logs
kubectl --namespace project-x logs statefulsets/cass-cc-1-cassandra-ringnodes

# Now we'll connect to Cassandra and run some queries
# (We'll use telepresence to allow us to run cqlsh as if we are in the project-x namespace)
telepresence --run-shell --namespace project-x

# There's a service which loadbalances CQL connections between all the Cassandra nodes.
cqlsh --cqlversion=3.4.2 cass-cc-1-cassandra-cql 9042

TRACING ON
SHOW VERSION
SHOW HOST
DESCRIBE CLUSTER

# Now let's scale up the cluster
# We use Helm to increment the replica count
helm upgrade \
     --debug \
     "cc-1" \
     contrib/charts/cassandra \
     --set pilotImage.repository=gcr.io/jetstack-sandbox/navigator-pilot-cassandra \
     --set replicaCount=2

# Run the node tool to see the cluster status
kubectl --namespace project-x exec cass-cc-1-cassandra-ringnodes-0 -- nodetool status


-- Database example is copied from:
-- http://blog.mclaughlinsoftware.com/2017/07/30/cassandra-query-language/
-- Create a database
DROP KEYSPACE IF EXISTS student;

CREATE KEYSPACE IF NOT EXISTS student
  WITH REPLICATION = {
     'class':'SimpleStrategy'
    ,'replication_factor': 2 }
  AND DURABLE_WRITES = true;

USE student;

-- See the ranges distributed to both nodes
DESCRIBE CLUSTER

-- Create some data
CONSISTENCY ALL

DROP TABLE IF EXISTS member;

CREATE TABLE member
( member_number       VARCHAR
, member_type         VARCHAR
, credit_card_number  VARCHAR
, credit_card_type    VARCHAR
, PRIMARY KEY ( member_number ));

DROP TABLE IF EXISTS contact;

CREATE TABLE contact
( contact_number      VARCHAR
, contact_type        VARCHAR
, first_name          VARCHAR
, middle_name         VARCHAR
, last_name           VARCHAR
, member_number       VARCHAR
, PRIMARY KEY ( contact_number ));

INSERT INTO member
( member_number, member_type, credit_card_number, credit_card_type )
VALUES
('SFO-12345','GROUP','2222-4444-5555-6666','VISA');

INSERT INTO contact
( contact_number, contact_type, first_name, middle_name, last_name, member_number )
VALUES
('CUS_00001','FAMILY','Barry', NULL,'Allen','SFO-12345');

INSERT INTO contact
( contact_number, contact_type, first_name, middle_name, last_name, member_number )
VALUES
('CUS_00002','FAMILY','Iris', NULL,'West-Allen','SFO-12345');

INSERT INTO member
( member_number, member_type, credit_card_number, credit_card_type )
VALUES
('SFO-12346','GROUP','3333-8888-9999-2222','VISA');

INSERT INTO contact
( contact_number, contact_type, first_name, middle_name, last_name, member_number )
VALUES
('CUS_00003','FAMILY','Caitlin','Marie','Snow','SFO-12346');

SELECT * FROM member;

# We can simulate an unresponsive node and see how the Readiness and Liveness probes detect the problem and restart the node.
# Stop one of the cassandra processes.
kubectl --namespace project-x  exec cass-cc-1-cassandra-ringnodes-0 -- bash -c 'kill -SIGSTOP -- $(ps -u cassandra -o pid=)'

# Nodetool now reports a node Down
kubectl --namespace project-x  exec cass-cc-1-cassandra-ringnodes-1 -- nodetool status

# And kubernetes reports that one of the pods is "unready"
kubectl --namespace project-x get pods

# Cleanup
helm delete --purge cc-1 navigator
kubectl delete ns --now project-x
