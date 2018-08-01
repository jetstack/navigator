Development guide
=================

Setting up
----------

Install minikube and start a cluster::

    minikube start --memory=8192 --cpus 4

Fetch the docker configuration::

    eval $(minikube docker-env)

Developing
----------

Edit code, then build::

    make build docker_build

Or only for the component you're interested in::

     make controller docker_build_controller

Testing
-------

Run
