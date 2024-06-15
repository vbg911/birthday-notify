package handlers

import (
	"birthday-notify/internal/config"
	"birthday-notify/internal/models"
	"birthday-notify/internal/storage"
	"birthday-notify/internal/storage/userRepo"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"time"
)

type UserHandler struct {
	userRepo storage.UserRepository
}

func NewUserHandler(repo storage.UserRepository) *UserHandler {
	return &UserHandler{userRepo: repo}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	_, err := mail.ParseAddress(user.Email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errors.New("invalid email")})
	}

	if len(user.Password) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password to short"})
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	user.Password = string(bytes)

	err = h.userRepo.CreateUser(user)
	if err != nil {
		if err == userRepo.ErrUserAlreadyExist {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user created"})
}

func (h *UserHandler) Subscribe(c *fiber.Ctx) error {
	var input models.UserSubscribe
	err := c.BodyParser(&input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	_, err = mail.ParseAddress(input.TargetEmail)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "incorrect email"})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	if input.TargetEmail == email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "can't subscribe yourself"})
	}

	err = h.userRepo.AddSubscriber(input.TargetEmail, email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "subscribe created"})
}

func (h *UserHandler) UnSubscribe(c *fiber.Ctx) error {
	var input models.UserSubscribe
	err := c.BodyParser(&input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	_, err = mail.ParseAddress(input.TargetEmail)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "incorrect email"})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	if input.TargetEmail == email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "can't unsubscribe from yourself"})
	}

	err = h.userRepo.DeleteSubscriber(input.TargetEmail, email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unsubscribe successfully"})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var input models.LoginUser

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.userRepo.GetUser(input.Email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "incorrect email or password"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "incorrect email or password"})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["birthday"] = user.Birthday
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	t, err := token.SignedString([]byte(config.GetJWTSecret()))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"message": "Success login", "token": t})
}
