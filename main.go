package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/nfnt/resize"

	_ "github.com/go-sql-driver/mysql"

	"k8s.io/client-go/rest"

	rdsv1alpha1clientset "github.com/eumel8/otc-rds-operator/pkg/rds/v1alpha1/apis/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Employee struct {
	Id    int
	Name  string
	City  string
	Photo string
}

var tmpl = template.Must(template.ParseGlob("form/*"))
var dataDir = os.Getenv("DATA_DIR")

func dbConn(w http.ResponseWriter) (db *sql.DB, err error) {

	rdsname := os.Getenv("MYSQL_NAME")
	namespace := os.Getenv("MYSQL_NAMESPACE")

	var dbUser string
	var dbPass string

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("KAUTH: " + err.Error())
		return nil, err
	}
	rdsclientset, err := rdsv1alpha1clientset.NewForConfig(restConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("KAUTH: " + err.Error())
		return nil, err
	}

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

			createTable := `CREATE TABLE IF NOT EXISTS employee (
      				id int(6) unsigned NOT NULL AUTO_INCREMENT,
				name varchar(30) NOT NULL,
 				city varchar(30) NOT NULL,
				photo varchar(128) NOT NULL,
				PRIMARY KEY (id)
				) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=latin1`

			_, err := db.Query(createTable)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("CREATETABLE: " + err.Error())
				return nil, err
			}

			return db, nil
		}
	}
}

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
}

func Show(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("DB: " + err.Error())
		return
	}
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM employee WHERE id=?", nId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("SHOW: " + err.Error())
		return
	}
	emp := Employee{}
	for selDB.Next() {
		var id int
		var name, city, photo string
		err = selDB.Scan(&id, &name, &city, &photo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("SHOW: " + err.Error())
			return
		}
		emp.Id = id
		emp.Name = name
		emp.City = city
		emp.Photo = photo
	}
	tmpl.ExecuteTemplate(w, "Show", emp)
	defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("DB: " + err.Error())
		return
	}
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM employee WHERE id=?", nId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("EDIT: " + err.Error())
		return
	}
	emp := Employee{}
	for selDB.Next() {
		var id int
		var name, city, photo string
		err = selDB.Scan(&id, &name, &city, &photo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("EDIT: " + err.Error())
			return
		}
		emp.Id = id
		emp.Name = name
		emp.City = city
		emp.Photo = photo
	}
	tmpl.ExecuteTemplate(w, "Edit", emp)
	defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("DB: " + err.Error())
		return
	}
	if r.Method == "POST" {
		name := r.FormValue("name")
		if name == "" {
			name = "none"
		}
		city := r.FormValue("city")
		if city == "" {
			city = "none"
		}
		photo := "none"
		_, handler, _ := r.FormFile("file")
		if handler != nil && handler.Header.Get("Content-Type") == "image/png" {
			uploadFile(w, r)
			photo = handler.Filename
		}
		insForm, err := db.Prepare("INSERT INTO employee(name, city, photo) VALUES(?,?,?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("INSERT: " + err.Error())
			return
		}
		insForm.Exec(name, city, photo)
		log.Println("INSERT: Name: " + name + " | City: " + city + " | Photo: " + photo)
	}
	defer db.Close()
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("UPLOAD: " + err.Error())
		return
	}

	defer file.Close()
	// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	// fmt.Printf("File Size: %+v\n", handler.Size)
	// fmt.Printf("MIME Header: %+v\n", handler.Header)

	dst, err := os.Create(dataDir + "/" + handler.Filename)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("UPLOAD: create " + err.Error())
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("UPLOAD: copy " + err.Error())
	}
	return
}

func Update(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("DB: " + err.Error())
		return
	}
	if r.Method == "POST" {
		name := r.FormValue("name")
		city := r.FormValue("city")
		photo := r.FormValue("photo")
		id := r.FormValue("uid")
		insForm, err := db.Prepare("UPDATE employee SET name=?, city=?, photo=? WHERE id=?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("UPDATE: " + err.Error())
			return
		}
		insForm.Exec(name, city, photo, id)
		log.Println("UPDATE: Name: " + name + " | City: " + city + " | Photo: " + photo)
	}
	defer db.Close()
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("DB: " + err.Error())
		return
	}
	emp := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM employee WHERE id=?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("DELETE: " + err.Error())
		return
	}
	delForm.Exec(emp)
	log.Println("DELETE")
	defer db.Close()
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

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
