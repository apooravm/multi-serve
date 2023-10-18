package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"encoding/json"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file")
	}
	PORT := os.Getenv("PORT")
	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/api/cronping", func(c echo.Context) error {
		return c.String(http.StatusOK, "sup")
	})

	e.GET("/pathParams/:name", getPathParams)
	e.GET("/queryParams/:name", getQueryParams)
	e.POST("/user", postUser)
	e.Static("/", "public")

	fmt.Printf("Live on %v", PORT)
	e.Logger.Fatal(e.Start(":" + PORT))

	readJsonFile("./db/users.json")
}

// /pathParams/:name
func getPathParams(c echo.Context) error {
	name := c.Param("name")
	fmt.Println(name)
	return c.String(http.StatusOK, name)
}

func getQueryParams(c echo.Context) error {
	allParams := c.QueryParams()
	fmt.Println("All Params:", allParams)

	colour := c.QueryParam("colour")
	fmt.Println("Colour:", colour)
	return c.String(http.StatusOK, "bruh")
}

func postUser(c echo.Context) error {
	newUser := new(User)
	if err := c.Bind(newUser); err != nil {
		return err
	}
	fmt.Println(newUser)
	return c.JSON(http.StatusCreated, newUser)
}

func readJsonFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error Opening %v", err)
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading the file %v", err)
	}

	var items []User

	if err := json.Unmarshal(content, &items); err != nil {
		log.Fatalf("Error unmarshalling: %v", err)
		return
	}

	fmt.Println("Items: ", items)

	items = append(items, User{"mrBruh", 39})

	updatedData, err := json.MarshalIndent(items, "", " ")
	if err != nil {
		log.Fatalf("Error Marshalling %v", err)
		return
	}

	if err := os.WriteFile(path, updatedData, 0644); err != nil {
		log.Fatalf("Error Writing to File %v", err)
		return
	}

	// fmt.Println(content)

	// var items []map[string]interface{}
	// if err := json.Unmarshal(content, &items); err != nil {
	// 	fmt.Println("Error unmarshaling JSON:", err)
	// 	return
	// }

	// // fmt.Println(items)
	// for key, value := range items {
	// 	fmt.Println(key, value)
	// }

}
