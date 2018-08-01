# Create a cluster
gcloud container clusters create --cluster-version 1.9.3-gke.0 --machine-type n1-standard-4 qcon

# Setup Helm/Tiller
kubectl create serviceaccount -n kube-system tiller
# Bind the tiller service account to the cluster-admin role
kubectl create clusterrolebinding tiller-binding --clusterrole=cluster-admin --serviceaccount kube-system:tiller
# Deploy tiller
helm init --service-account tiller


# Install Navigator
helm install --name nav ./contrib/charts/navigator --set apiserver.persistence.enabled=true --namespace navigator

# Create storageclasses in each zone
kubectl create -f sc-a.yaml
kubectl create -f sc-b.yaml

# Create a cassandra cluster
kubectl create -f cassandra-cluster.yaml

kubectl port-forward to grafana (port 3000)

Import custom dashboard JSON in this directory if required

View dashboard

Exec into a couple of cassandra pods, run

kubectl exec -it cass-demo-europe-west2-a-0 bash
unset JVM_OPTS
cassandra-stress write duration=30s -node cass-demo-nodes

View dashboard
