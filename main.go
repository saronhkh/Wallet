package main

import (
	db "Wallet/config"
	routes "Wallet/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {

	db.Connect()
	//create instsnce of fiber
	app := fiber.New()

	routes.Setup(app)

	//listen on port
	app.Listen(":3000")
}
