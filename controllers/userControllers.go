package controllers

import (
	db "Wallet/config"
	models "Wallet/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

func CreateUser(c *fiber.Ctx) error {
	var data map[string]string

	err := c.BodyParser(&data)
	if err != nil {
		return c.Status(400).JSON(
			fiber.Map{
				"success": false,
				"message": "invalid data",
			})
	}

	validate := validator.New()

	user := models.User{
		Name:      data["name"],
		Email:     data["email"],
		Password:  data["password"],
		CreatedAt: time.Now(),
	}

	errval := validate.Struct(user)
	if validationErrors, ok := errval.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, err := range validationErrors {
			field := err.Field()
			tag := err.Tag()

			var message string
			switch tag {
			case "required":
				message = field + " This field is required."
			case "email":
				message = field + " Must be a valid email address."
			case "min":
				message = field + " Must be at least 8 characters long"
			default:
				message = "Validation error"
			}

			errors[field] = message
		}

		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"errors":  errors,
		})
	}

	hashedPassword, errpass := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if errpass != nil {
		return c.Status(400).JSON(
			fiber.Map{
				"success": false,
				"message": errpass.Error(),
			})
	}

	user.Password = string(hashedPassword)

	db.DB.Create(&user)
	wallet_id := user.User_id + 1000000000

	wallet := models.Wallet{
		Wallet_id: wallet_id,
		Balance:   0,
		CreatedAt: time.Now(),
		UserRefer: user.User_id,
		User:      user,
	}
	db.DB.Create(&wallet)

	return c.Status(200).JSON(
		fiber.Map{
			"success": true,
			"message": "user and wallet created successfully",
			"data":    wallet,
		})
}

func LogIn(c *fiber.Ctx) error {

	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var loginReq LoginRequest

	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(400).JSON(
			fiber.Map{
				"success": false,
				"message": "invalid data, email and password required",
			})
	}

	var user models.User
	if err := db.DB.Where("email = ?", loginReq.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"success": false,
				"message": "User not found",
			})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"success": false,
				"message": "Incorrect password",
			})
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.User_id
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix() // 10 minutes.

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return c.Status(401).JSON(
			fiber.Map{
				"success": false,
				"message": err.Error(),
			})
	}

	return c.Status(200).JSON(
		fiber.Map{
			"success": true,
			"message": "Login successful",
			"user":    user,
			"token":   tokenString,
		})

}

func WalletDetails(c *fiber.Ctx) error {

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	user_id := claims["user_id"]

	var wallet models.Wallet
	db.DB.Where("user_refer = ?", user_id).Preload("User").First(&wallet)

	return c.Status(200).JSON(
		fiber.Map{
			"success": true,
			"wallet":  wallet,
		})
}

func AddFunds(c *fiber.Ctx) error {

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	user_id := claims["user_id"]

	var input struct {
		Amount float64 `json:"amount" validate:"required,min=0.5"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input"})
	}

	validate := validator.New()
	errval := validate.Struct(input)
	if validationErrors, ok := errval.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, err := range validationErrors {
			field := err.Field()
			tag := err.Tag()

			var message string
			switch tag {
			case "required":
				message = field + " This field is required."
			case "min":
				message = field + " The amount must be at least 0.5. "
			default:
				message = "Validation error"
			}

			errors[field] = message
		}

		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"errors":  errors,
		})
	}

	var wallet models.Wallet
	db.DB.Where("user_refer = ?", user_id).Preload("User").First(&wallet)

	wallet.Balance += input.Amount

	if err := db.DB.Save(&wallet).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Could not update wallet"})
	}

	return c.Status(200).JSON(
		fiber.Map{
			"success": true,
			"message": "Add the amount successfully",
			"wallet":  wallet,
		})
}

func TransferFunds(c *fiber.Ctx) error {

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	user_id := claims["user_id"]

	var sourwallet, destwallet models.Wallet
	db.DB.Where("user_refer = ?", user_id).First(&sourwallet)

	var transferRequest struct {
		ToWalletID uint    `json:"to_wallet_id" validate:"required"`
		Amount     float64 `json:"amount" validate:"required,min=0.5"`
	}

	if err := c.BodyParser(&transferRequest); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input"})
	}

	validate := validator.New()
	errval := validate.Struct(transferRequest)
	if validationErrors, ok := errval.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, err := range validationErrors {
			field := err.Field()
			tag := err.Tag()

			var message string
			switch tag {
			case "required":
				message = field + " This field is required."
			case "min":
				message = field + " The amount must be at least 0.5. "
			default:
				message = "Validation error"
			}

			errors[field] = message
		}

		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"errors":  errors,
		})
	}

	if err := db.DB.Where("wallet_id = ?", transferRequest.ToWalletID).First(&destwallet).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": " destination wallet id not found"})
	}

	if sourwallet.Balance < transferRequest.Amount {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": " insufficient funds "})
	}

	sourwallet.Balance -= transferRequest.Amount
	destwallet.Balance += transferRequest.Amount

	if err := db.DB.Save(&sourwallet).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Could not update wallet"})
	}

	if err := db.DB.Save(&destwallet).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Could not update wallet"})
	}

	transaction := models.Transaction{
		Source_wallet_id:      sourwallet.Wallet_id,
		Destination_wallet_id: destwallet.Wallet_id,
		Amount:                transferRequest.Amount,
		Timestamp:             time.Now(),
	}

	if err := db.DB.Create(&transaction).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Could not create wallet"})
	}

	return c.Status(200).JSON(
		fiber.Map{
			"success":                 true,
			"message":                 "transfer the amount successfully",
			"transaction":             transaction,
			"Destination new Balance": destwallet.Balance,
			"Source new Balance":      sourwallet.Balance,
		})
}

func TransactionHistory(c *fiber.Ctx) error {

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	user_id := claims["user_id"]

	var wallet models.Wallet
	db.DB.Where("user_refer = ?", user_id).Select("wallet_id").First(&wallet)
	wallet_id := wallet.Wallet_id

	var transactions []models.Transaction
	if err := db.DB.Where("source_wallet_id = ? OR destination_wallet_id = ?", wallet_id, wallet_id).Find(&transactions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "could not retrieve transactions",
		})
	}

	return c.Status(200).JSON(
		fiber.Map{
			"success":              true,
			"for wallet_id":        wallet_id,
			"transactions history": transactions,
		})
}
