package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lib/pq"
)

type Expense struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Amount float64  `json:"amount"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
}

type Err struct {
	Message string `json:"message"`
}

func createExpenseHandler(c echo.Context) error {
	ex := Expense{}
	err := c.Bind(&ex)

	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	row := db.QueryRow("INSERT INTO expenses (title, amount, note, tags ) values ($1, $2, $3, $4) RETURNING id,title, amount, note, tags", ex.Title, ex.Amount, ex.Note, pq.Array(ex.Tags))
	err = row.Scan(&ex.ID, &ex.Title, &ex.Amount, &ex.Note, pq.Array(&ex.Tags))

	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, ex)
}

func getAllExpenseHandler(c echo.Context) error {
	stmt, err := db.Prepare("SELECT id, title, amount, note, tags  FROM expenses")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expenses statment:" + err.Error()})
	}

	row, err := stmt.Query()
	expenses := []Expense{}

	for row.Next() {
		ex := Expense{}
		err := row.Scan(&ex.ID, &ex.Title, &ex.Amount, &ex.Note, pq.Array(&ex.Tags))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expense:" + err.Error()})
		}
		expenses = append(expenses, ex)
	}

	return c.JSON(http.StatusOK, expenses)
}

func getExpenseHandler(c echo.Context) error {
	id := c.Param("id")
	stmt, err := db.Prepare("SELECT id, title, amount, note, tags  FROM expenses where id = $1")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expenses statment:" + err.Error()})
	}

	row := stmt.QueryRow(id)
	ex := Expense{}
	err = row.Scan(&ex.ID, &ex.Title, &ex.Amount, &ex.Note, pq.Array(&ex.Tags))
	switch err {
	case sql.ErrNoRows:
		return c.JSON(http.StatusNotFound, Err{Message: "expense not found"})
	case nil:
		return c.JSON(http.StatusOK, ex)
	default:
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expense:" + err.Error()})
	}
}

func putExpenseHandler(c echo.Context) error {
	id := c.Param("id")
	ex := Expense{}
	err := c.Bind(&ex)

	stmt, err := db.Prepare("UPDATE expenses SET  title = $2, amount = $3, note = $4, tags = $5 where id = $1 RETURNING id,title, amount,note,tags")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expenses statment:" + err.Error()})
	}

	row := stmt.QueryRow(id, ex.Title, ex.Amount, ex.Note, pq.Array((ex.Tags)))

	err = row.Scan(&ex.ID, &ex.Title, &ex.Amount, &ex.Note, pq.Array(&ex.Tags))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expense:" + err.Error()})
	}
	return c.JSON(http.StatusOK, ex)
}

var db *sql.DB

func main() {

	fmt.Println("Please use server.go for main file")
	var err error
	fmt.Println("start at port:", os.Getenv("PORT"))
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database error", err)
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF NOT EXISTS expenses (id SERIAL PRIMARY KEY , title TEXT , amount FLOAT , note TEXT ,tags TEXT[]);
	`
	_, err = db.Exec(createTb)

	if err != nil {
		log.Fatal("Can't create table", err)
	}
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/expenses", createExpenseHandler)
	e.GET("/expenses/:id", getExpenseHandler)
	e.GET("/expenses", getAllExpenseHandler)
	e.PUT("/expenses/:id", putExpenseHandler)

	log.Fatal(e.Start(":2566"))
}
