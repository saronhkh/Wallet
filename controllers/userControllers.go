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
