Demoapp with Go MySQL
=====================

This demo app can be deploy on a Kubernetes cluster and will use
the following feature:

* Webapplication with UI
* MySQL backend (e.g. RDS OTC)
* Storage (requires default StorageClass in cluster)

Usage
-----

Adjust credentials and connection in `demoapp.yaml` for database connection

```bash
MYSQL_USER # mysql username, e.g. `app` or `root` in init container
MYSQL_PASSWORD # mysql password for `app`or `root` in init container
MYSQL_HOST # mysql hostname
MYSQL_PORT # mysql port, e.g. 3306
MYSQL_DB # mysql db, e.g. `app`
DATA_DIR # location of app images, e.g `/data`
```
and deploy manifest:

```bash
kubectl apply -f demoapp.yaml
```

If the cluster supports Ingress, adjust `demoapp-ingress.yaml` (e.g. hostname)
and deploy manifest:

```bash
kubectl apply -f demoapp-ingress.yaml
```
