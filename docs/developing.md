Quickstart for developing Navigator
===================================

Setting up
----------

Install minikube and start a cluster with RBAC enabled:

    minikube start --extra-config=apiserver.Authorization.Mode=RBAC

Work around `kube-dns` and helm having problems when RBAC is enabled in minikube:

    kubectl create clusterrolebinding cluster-admin:kube-system \
        --clusterrole=cluster-admin \
        --serviceaccount=kube-system:default

Fetch the docker configuration:

    eval $(minikube docker-env)

Build images in minikube's docker:

    make BUILD_TAG=dev all

Or quicker (skips tests):

    make BUILD_TAG=dev build docker_build

Install helm into the minikube cluster:

    helm init

Install navigator using the helm chart:

    helm install contrib/charts/navigator \
        --set apiserver.image.pullPolicy=Never \
        --set apiserver.image.tag=dev \
        --set controller.image.pullPolicy=Never \
        --set controller.image.tag=dev \
        --name navigator --namespace navigator --wait

Now test navigator is deployed by creating a demo elasticsearch cluster. Edit
`docs/quick-start/es-cluster-demo.yaml` to change the pilot image tag to `dev`,
and set the `pullPolicy` to `Never`, then create the cluster:

    kubectl create -f docs/quick-start/es-cluster-demo.yaml


Developing
----------

Edit code, then build:

    make BUILD_TAG=dev build docker_build

Or only for the component you're interested in:

     make BUILD_TAG=dev controller docker_build_controller

Kill the component you're working on, for example the controller:

    kubectl delete pods -n navigator -l app=navigator -l component=controller
