package main

import (
	"fmt"

	"github.com/rusq/goflick"
)

func main() {
	flick, err := goflick.NewConnect("email", "password")
	if err != nil {
		panic(err)
	}
	// get current price
	price, err := flick.GetPrice()
	if err != nil {
		panic(err)
	}
	fmt.Println(price)
}
