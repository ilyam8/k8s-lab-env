# k8s-lab-env

This is my lab for testing auto-discovery in [Kubernetes](https://kubernetes.io/).

`k8s/env` dir contains application deployments scripts.

### deploy

> ./k8s/install_env.sh.

It creates `netdata-auto-discovery` namespace and installs bunch of applications into it.

### delete

> kubectl delete namespace netdata-auto-discovery
 

### bouncer

Is a simple tool that periodically changes number of deployment replicas in the `netdata-auto-discovery` namespace.

Keep in mind that lab environment should be [created](#deploy) first.
