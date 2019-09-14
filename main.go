package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func init() {
	db := setupDatabase()
	defer db.Close()
	createCustomerTable(db)
}

type customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func setupDatabase() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database error", err)
	}
	fmt.Println("okey")
	return db
}

func authMiddleware(c *gin.Context) {
	log.Println("start middleware")
	authKey := c.GetHeader("Authorization")
	if authKey != "token2019" {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		c.Abort()
		return
	}
}

func insertCustomerHandler(c *gin.Context) {
	cus := customer{}
	if err := c.ShouldBindJSON(&cus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := insertCustomerTable(cus.Name, cus.Email, cus.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cus, err = getCustomerByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cus)
}

func getCustomerByIDHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cus, err := getCustomerByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cus)
}

func getCustomerHandler(c *gin.Context) {
	cus, err := getCustomer()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cus)
}

func updateCustomerHandler(c *gin.Context) {
	cus := customer{}
	if err := c.ShouldBindJSON(&cus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := updateCustomer(cus)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cus, err = getCustomerByID(cus.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cus)
}

func deleteCustomerByIDHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = deleteTodosStatusByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func createCustomerTable(db *sql.DB) {
	createTb := `CREATE TABLE IF NOT EXISTS customers (id SERIAL PRIMARY KEY, name TEXT, email TEXT, status TEXT);`
	_, err := db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table ", err)
	}
	fmt.Println("create table success")
}

func insertCustomerTable(name, email, status string) (int, error) {
	insertTb := `INSERT INTO customers (name, email, status) values ($1, $2, $3) RETURNING id`
	db := setupDatabase()
	defer db.Close()
	row := db.QueryRow(insertTb, name, email, status)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return id, err
	}
	return id, err
}

func getCustomerByID(id int) (customer, error) {
	selectTb := `SELECT id, name, email, status FROM customers where id=$1`
	cus := customer{}
	db := setupDatabase()
	defer db.Close()
	stmt, err := db.Prepare(selectTb)
	if err != nil {
		return cus, err
	}
	row := stmt.QueryRow(id)
	err = row.Scan(&cus.ID, &cus.Name, &cus.Email, &cus.Status)
	if err != nil {
		return cus, err
	}
	return cus, err
}

func getCustomer() ([]customer, error) {
	selectTb := `SELECT id, name, email, status FROM customers`
	db := setupDatabase()
	defer db.Close()
	cuss := []customer{}
	stmt, err := db.Prepare(selectTb)
	if err != nil {
		return cuss, err
	}
	rows, err := stmt.Query()
	if err != nil {
		return cuss, err
	}
	for rows.Next() {
		cus := customer{}
		err := rows.Scan(&cus.ID, &cus.Name, &cus.Email, &cus.Status)
		if err != nil {
			return cuss, err
		}
		cuss = append(cuss, cus)
	}
	return cuss, err
}

func updateCustomer(cus customer) error {
	updateTb := `UPDATE customers SET name=$2, email=$3, status=$4 WHERE id=$1;`
	db := setupDatabase()
	defer db.Close()
	stmt, err := db.Prepare(updateTb)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(cus.ID, cus.Name, cus.Email, cus.Status); err != nil {
		return err
	}
	fmt.Println("update success")
	return err
}

func deleteTodosStatusByID(id int) error {
	deleteTb := `DELETE FROM customers WHERE id=$1;`
	db := setupDatabase()
	defer db.Close()
	stmt, err := db.Prepare(deleteTb)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(id); err != nil {
		return err
	}
	fmt.Println("delete success")
	return err
}

func main() {
	r := setupRouter()
	r.Use(authMiddleware)
	r.POST("/customers", insertCustomerHandler)
	r.GET("/customers/:id", getCustomerByIDHandler)
	r.GET("/customers", getCustomerHandler)
	r.PUT("/customers/:id", updateCustomerHandler)
	r.DELETE("/customers/:id", deleteCustomerByIDHandler)
	r.Run(":2019")
}
