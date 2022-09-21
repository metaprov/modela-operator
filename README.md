# Modela Operator

Modela is a Kubernetes-native machine learning platform with an emphasis on AutoML and MLOps. The Modela Operator
supports deploying Modela and is designed to quickly bootstrap an operational installation.


## Description

The Modela Operator runs as a container on your Kubernetes cluster and provides logic for the `Modela` Custom Resource.
The resource defines the configuration for prerequisite components and Modela Tenants. Prerequisites include:

* Cert Manager
* PostgreSQL
* MinIO (Optional)
* NGINX Ingress Controller (Optional)
* Prometheus, Loki, Grafana (Optional)


## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [Rancher Desktop](https://rancherdesktop.io/), [Docker Desktop](https://www.docker.com/products/docker-desktop/), or [KIND](https://sigs.k8s.io/kind) to run a cluster on your local machine.

### YAML Install

1. Install the Modela Operator

```sh
kubectl create -k "https://github.com/metaprov/modela-operator/config/default" 
```

2. Apply the Modela Custom Resource

| Sample File | Installation Type |
| --- | ----------- |
| modela_base.yaml | Required prerequisites only |
| modela_full.yaml | Required and optional prerequisites |
| modela_nginx_ingress.yaml | Install & configure NGINX (recommended)|

```sh
kubectl create -f "https://raw.githubusercontent.com/metaprov/modela-operator/main/config/samples/modela_nginx_ingress.yaml" 
```

### Access Modela

Once the Modela Custom Resource has been created the Modela Operator will install Modela in around 3-5 minutes. Once the installation is complete you will be able to access the Modela dashboard 
at [http://modela-app.localhost](http://modela-app.localhost). If you did not enable Ingress as part of the installation you
will need to use the [Modela CLI](https://modela.ai/docs/docs/install/modela/quick/) to access the dashboard.

The default access credentials after installation are as follows:
* Tenant Name - `modela`
* Account Name & Password - `admin`


## License

This project is licensed under the Apache-2.0 License.