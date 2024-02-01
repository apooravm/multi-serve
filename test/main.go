package main

import (
	"fmt"
)

type Person struct {
	Name     string
	FavSport string
}

func main() {
	randMap := make(map[string]string)
	randMap["a"] = "A"
	randMap["b"] = "B"
	randMap["c"] = "C"

	// if randMap["d"] == "" {
	// 	fmt.Println("Nothin")
	// }

	personMap := make(map[string]*Person)
	personMap["a"] = &Person{
		Name:     "MrBruh",
		FavSport: "Bruhket",
	}

	val := personMap["v"]
	fmt.Printf("this is extra %v", nil)

	fmt.Printf("type %T\n", val)
	if val == nil {
		fmt.Println("test/main.go ln:33 |", "Is nil")
	}
}
