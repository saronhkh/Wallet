package routes

import (
	controllers "Wallet/controllers"
	"os"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

func Setup(app *fiber.App) {

	//user routes
	app.Post("/signup", controllers.CreateUser)
	app.Post("/login", controllers.LogIn)

	jwtMiddleware := jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	})

	protected := app.Group("/wallet")
	protected.Use(jwtMiddleware)
	protected.Get("", controllers.WalletDetails)
	protected.Post("/add-funds", controllers.AddFunds)
	protected.Post("/transfer", controllers.TransferFunds)
	protected.Get("/transactions", controllers.TransactionHistory)
}
