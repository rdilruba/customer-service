package service

import (
	"customer-service/db"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Customer struct {
	ID      int    `json:"id"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email"`
	Address string `json:"address,omitempty"`
}

func createCustomer(db *db.PostgresDB, c *gin.Context) (int, *Customer, error) {
	var customer Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		return http.StatusBadRequest, nil, err
	}

	if len(customer.Email) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("email cannot be empty")
	}
	stmt := `INSERT INTO customers (name, email, address) VALUES ($1, $2, $3) RETURNING id`
	err := db.DB.QueryRow(stmt, customer.Name, customer.Email, customer.Address).Scan(&customer.ID)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusCreated, &customer, nil
}

func getCustomer(db *db.PostgresDB, c *gin.Context) (int, *Customer, error) {
	customerID := c.Param("customerId")
	id, err := strconv.Atoi(customerID)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	var customer Customer
	stmt := `SELECT id, name, email, address FROM customers WHERE id = $1`
	err = db.DB.QueryRow(stmt, id).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Address)

	if err == sql.ErrNoRows {
		return http.StatusNotFound, nil, err
	}

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, &customer, nil
}

func updateCustomer(db *db.PostgresDB, c *gin.Context) (int, *Customer, error) {
	customerID := c.Param("customerId")
	id, err := strconv.Atoi(customerID)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	var customer Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		return http.StatusBadRequest, nil, err
	}

	fieldsNum := 0
	fields := make([]interface{}, 0)
	stmt := `UPDATE customers SET `
	if len(customer.Address) != 0 {
		fieldsNum += 1
		stmt += fmt.Sprintf("address = $%d ", fieldsNum)
		fields = append(fields, customer.Address)
	}
	if len(customer.Name) != 0 {
		fieldsNum += 1
		stmt += fmt.Sprintf("name = $%d ", fieldsNum)
		fields = append(fields, customer.Name)
	}
	if fieldsNum == 0 {
		return http.StatusNotModified, &customer, nil
	}
	fieldsNum += 1
	stmt += fmt.Sprintf("WHERE id = $%d RETURNING id, name, email, address", fieldsNum)
	fields = append(fields, id)

	err = db.DB.QueryRow(stmt, fields...).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Address)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	customer.ID = id
	return http.StatusOK, &customer, nil
}

func deleteCustomer(db *db.PostgresDB, c *gin.Context) (int, error) {
	customerID := c.Param("customerId")
	id, err := strconv.Atoi(customerID)
	if err != nil {
		return http.StatusBadRequest, err
	}

	stmt := `DELETE FROM customers WHERE id = $1`
	_, err = db.DB.Exec(stmt, id)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusNoContent, nil
}
