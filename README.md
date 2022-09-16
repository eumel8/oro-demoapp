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

Go:

```go
import (
	// ...
	"k8s.io/client-go/rest"
	rdsv1alpha1clientset "github.com/eumel8/otc-rds-operator/pkg/rds/v1alpha1/apis/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
```

Import Kubernetes Go Rest Client, OTC RDS API clientset, and Kubernetes API Metadata

```go
	// ...
	for {
		log.Println("LOOP: get rds " + rdsname)
	
	rds, err := rdsclientset.McspsV1alpha1().Rdss(namespace).Get(context.TODO(), rdsname, metav1.GetOptions{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("KRDS: " + err.Error())
			return nil, err
		}
		if rds.Status.Status == "ACTIVE" {
			log.Println("DB ACTIVE")
			for _, i := range *rds.Spec.Users {
				dbUser = i.Name
				dbPass = i.Password
				break
			}
			dbDriver := "mysql"
			dbHost := rds.Status.Ip
			dbPort := rds.Spec.Port
			dbName := rds.Spec.Databases[0]
			log.Println("DB CONNECT " + dbHost)
			db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbHost+":"+dbPort+")/"+dbName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("DBCONN: " + err.Error())
				return nil, err
			}
		}
		// ...
		time.Sleep(5 * time.Second)
	}
```

After in-cluster authentication on K8S API, create rdsclientset,
search for specific API and fetch, ip-address of RDS instance,
username/password of app user and try to connect

