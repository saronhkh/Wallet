package main

import (
	db "Wallet/config"
	routes "Wallet/routes"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {

	db.Connect()
	//create instsnce of fiber
	app := fiber.New()

	routes.Setup(app)

	//listen on port
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("faild to start server", err)
	}
}
