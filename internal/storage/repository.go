package storage

import (
	"birthday-notify/internal/models"
)

type UserRepository interface {
	CreateUser(user models.User) error
	GetUser(email string) (*models.User, error)
	AddSubscriber(userEmail string, subscriberEmail string) error
	FindBirthdayPeoples(day int, month int) ([]models.User, error)
	DeleteSubscriber(userEmail string, subscriberEmail string) error
}
