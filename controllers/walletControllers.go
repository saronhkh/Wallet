package controllers

import (
	db "Wallet/config"
	models "Wallet/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/go-playground/validator.v9"
)

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
