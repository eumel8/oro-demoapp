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

```go
func Index(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("DB: " + err.Error())
		return
	}
	selDB, err := db.Query("SELECT * FROM employee ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("INDEX: " + err.Error())
		return
	}
	emp := Employee{}
	res := []Employee{}
	for selDB.Next() {
		var id int
		var name, city, photo string
		err = selDB.Scan(&id, &name, &city, &photo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("INDEX 2: " + err.Error())
			return
		}
		emp.Id = id
		emp.Name = name
		emp.City = city
		if photo != "none" {
			f, err := os.Open(dataDir + "/" + photo)
			if err != nil {
				// http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("INDEX : photoload " + err.Error())
				// return
			} else {
				img, _, err := image.Decode(f)
				sane := resize.Resize(100, 100, img, resize.Bilinear)
				var buff bytes.Buffer
				png.Encode(&buff, sane)

				encodedString := base64.StdEncoding.EncodeToString(buff.Bytes())
				emp.Photo = encodedString
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Println("INDEX : photodecode" + err.Error())
					return
				}
			}
			defer f.Close()
		} else {
			emp.Photo = "iVBORw0KGgoAAAANSUhEUgAAAJoAAAB/CAYAAAAXdtsmAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAAFiUAABYlAUlSJPAAAAFdSURBVHhe7dKxAYAwDMCw0P9/Boa+EE/S4gf8vL+BZecWVhmNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjYTRSBiNhNFIGI2E0UgYjcDMB+WSBPrvm9bgAAAAAElFTkSuQmCC"
		}
		res = append(res, emp)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
```

Using [Go text/template package](https://pkg.go.dev/text/template) to provide HTML views
for the web app. Form data are stored in MySQL database. The user can upload a file,
which is stored on specific data dir.

```go
func main() {
	log.Println("Server started on: :8080")
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/new", New)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)
	http.ListenAndServe(":8080", nil)
}
```

Routing of http handler to different function. Service web app on port :8080

```yaml
---
apiVersion: apps/v1
kind: StatefulSet
spec:
  template:
    spec:
      containers:
      - image: ghcr.io/eumel8/demoapp:1.0.0
        name: demoapp
        ports:
        - containerPort: 8080
        env:
        - name: MYSQL_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.serviceAccountName
        - name: MYSQL_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: DATA_DIR
          value: "/data" # external pvc
        volumeMounts:
        - mountPath: /data
          name: demoapp-volume
      serviceAccountName: demoapp
      volumes:
      - name: demoapp-volume
        persistentVolumeClaim:
          claimName: demoapp-volume
```

Use K8S metadata for `MYSQL_NAME` and `MYSQL_NAMESPACE`,
Attach external PVC to the web app container, and expose containerPort 8080

```
FROM mtr.devops.telekom.de/mcsps/golang:1.18 as builder

WORKDIR /app
ADD . /app

RUN go mod download && go mod tidy && go build -o main main.go

FROM gcr.io/distroless/base
LABEL org.opencontainers.image.authors="f.kloeker@telekom.de"
LABEL version="1.0.0"
LABEL description="Create DemoApp for OTC RDS Operator"

WORKDIR /

USER nonroot:nonroot
COPY --from=builder --chown=nonroot:nonroot /app/main /
COPY --from=builder --chown=nonroot:nonroot /app/kodata /var/run/ko

ENV KO_DATA_PATH=/var/run/ko
EXPOSE 8080
ENTRYPOINT ["/main"]
```

A Dockerfile to compile to Go binary and copy the result in a distroless
images

```yaml
name: Build
# https://github.com/marketplace/actions/cosign-installer
on:
  push:
    branches:
      - "**"
  pull_request:
    branches: [ master ]
  release:
    types: [created]
    
env:
  IMAGE_NAME: oro-demoapp/demoapp
  
jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - name: Setup Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.18

    - name: Checkout Repo
      uses: actions/checkout@v2

    - name: Setup ko
      uses: imjasonh/setup-ko@v0.6

    - name: Install Cosign
      uses: sigstore/cosign-installer@main

    - name: Log in to ghcr registry
      run: |
        echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin
        echo "${{ secrets.GITHUB_TOKEN }}" | ko login ghcr.io -u ${{ github.actor }} --password-stdin

    - name: Build the image with ko
      run: |
        ko build -t ko --sbom none --bare
      env:
       KO_DOCKER_REPO: ghcr.io/${{ github.repository_owner }}/oro-demoapp/demoapp

    - name: Build the image with docker
      run: docker build . --file Dockerfile --tag $IMAGE_NAME --label "runnumber=${GITHUB_RUN_ID}"

    - name: Push & Sign image
      run: |
        IMAGE_ID=ghcr.io/${{ github.repository_owner }}/$IMAGE_NAME
        # Change all uppercase to lowercase
        IMAGE_ID=$(echo $IMAGE_ID | tr '[A-Z]' '[a-z]')
        # Strip git ref prefix from version
        VERSION=$(echo "${{ github.ref }}" | sed -e 's,.*/\(.*\),\1,')
        # Strip "v" prefix from tag name
        [[ "${{ github.ref }}" == "refs/tags/"* ]] && VERSION=$(echo $VERSION | sed -e 's/^v//')
        # Use Docker `latest` tag convention
        [ "$VERSION" == "master" ] && VERSION=latest
        echo IMAGE_ID=$IMAGE_ID
        echo VERSION=$VERSION
        echo "{{ github.ref.type }}"
        docker tag $IMAGE_NAME $IMAGE_ID:$VERSION
        docker push $IMAGE_ID:$VERSION
        cosign sign --key env://COSIGN_KEY $IMAGE_ID:$VERSION
        cosign sign --key env://COSIGN_KEY ghcr.io/${{ github.repository_owner }}/$IMAGE_NAME:ko
      env:
        COSIGN_KEY: ${{secrets.COSIGN_KEY}}
        COSIGN_PASSWORD: ${{secrets.COSIGN_PASSWORD}}

    - name: Push & Sign image release
      if: github.ref.type == 'tag'
      run: |
        IMAGE_ID=ghcr.io/${{ github.repository_owner }}/$IMAGE_NAME
        IMAGE_ID=$(echo $IMAGE_ID | tr '[A-Z]' '[a-z]')
        VERSION=${GITHUB_REF_NAME}
        docker tag $IMAGE_NAME $IMAGE_ID:$VERSION
        docker push $IMAGE_ID:$VERSION
        cosign sign --key env://COSIGN_KEY $IMAGE_ID:$VERSION
      env:
        COSIGN_KEY: ${{secrets.COSIGN_KEY}}
        COSIGN_PASSWORD: ${{secrets.COSIGN_PASSWORD}}
```

Github Action build the image with docker and with [ko](https://github.com/ko-build/ko)
which don't need a Dockerfile definition. Both images are signed with [cosign](https://github.com/sigstore/cosign)
 to verify with the [pubkey](https://raw.githubusercontent.com/eumel8/otc-rds-operator/master/cosign.pub)

Using the [Github Registry](https://github.com/eumel8/oro-demoapp/pkgs/container/oro-demoapp%2Fdemoapp)

## Credits

Frank Kloeker f.kloeker@telekom.de

Life is for sharing. If you have an issue with the code or want to improve it, feel free to open an issue or an pull request.

Go Crud Example is adapted from [www.golangprograms.com](https://www.golangprograms.com/example-of-golang-crud-using-mysql-from-scratch.html)
