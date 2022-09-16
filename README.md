OTC RDS Operator Demoapp
========================

![](oro-demoapp.png)

This demo is part of [OTC RDS Operator](https://github.com/eumel8/otc-rds-operator)
and deploy on a Kubernetes cluster and will use the following feature:

* Webapplication with UI
* MySQL backend (RDS OTC)
* Storage (requires default StorageClass in cluster)

Usage
-----

Ensure existing OTC VPC,Subnet, and SecurityGroup in `demoapp-rds.yaml`

and deploy manifest:

```bash
kubectl apply -f demoapp-rds.yaml
```

deploy manifest:

```bash
kubectl apply -f demoapp.yaml
```

If the cluster supports Ingress, adjust `demoapp-ingress.yaml` (e.g. hostname)
and deploy manifest:

```bash
kubectl apply -f demoapp-ingress.yaml
```

On Ingress endpoint in your browser should appear `Golang Mysql Curd Example`

Deep Dive
---------
