package userRepo

import (
	"birthday-notify/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"slices"
	"time"
)

type UserRepo struct {
	db *pgxpool.Pool
}

var (
	ErrUserAlreadyExist  = errors.New("user already exist")
	ErrNotExist          = errors.New("row does not exist")
	ErrAlreadySubscriber = errors.New("already subscribed")
	//ErrNoBirthdayPeoples = errors.New("birthday people doesnt exist at this date")
)

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) CreateUser(u models.User) error {
	query := "INSERT INTO public.users(email, password, birthday) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(context.Background(), query, u.Email, u.Password, u.Birthday)
	if err != nil {
		var pgxError *pgconn.PgError
		if errors.As(err, &pgxError) {
			if pgxError.Code == "23505" {
				return ErrUserAlreadyExist
			}
		}
		return err
	}
	return nil
}

func (r *UserRepo) GetUser(email string) (*models.User, error) {
	query := "SELECT id, email, password, birthday, subscribers FROM public.users WHERE email=$1"
	row := r.db.QueryRow(context.Background(), query, email)
	var user models.User
	var birthday time.Time
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &birthday, &user.Subscribers); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &user, ErrNotExist
		}
		return &user, err
	}

	user.Birthday = birthday.Format(time.DateOnly)

	return &user, nil
}

func (r *UserRepo) DeleteSubscriber(userEmail string, subscriberEmail string) error {
	var currentSubscribers []string

	// Чтение текущих подписчиков
	query := "SELECT subscribers FROM public.users WHERE email = $1"
	err := r.db.QueryRow(context.Background(), query, userEmail).Scan(&currentSubscribers)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user with email %s not found", userEmail)
		}
		return fmt.Errorf("failed to get current subscribers: %v", err)
	}

	index := slices.Index(currentSubscribers, subscriberEmail)
	if index != -1 {
		currentSubscribers = append(currentSubscribers[:index], currentSubscribers[index+1:]...)
	} else {
		return errors.New("you are not a subscriber")
	}

	// Обновление поля subscribers
	updateQuery := "UPDATE public.users SET subscribers = $1 WHERE email = $2"
	_, err = r.db.Exec(context.Background(), updateQuery, currentSubscribers, userEmail)
	if err != nil {
		return fmt.Errorf("failed to update subscribers: %v", err)
	}
	return nil
}

func (r *UserRepo) AddSubscriber(userEmail string, subscriberEmail string) error {
	var currentSubscribers []string

	// Чтение текущих подписчиков
	query := "SELECT subscribers FROM public.users WHERE email = $1"
	err := r.db.QueryRow(context.Background(), query, userEmail).Scan(&currentSubscribers)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user with email %s not found", userEmail)
		}
		return fmt.Errorf("failed to get current subscribers: %v", err)
	}

	if slices.Contains(currentSubscribers, subscriberEmail) {
		return ErrAlreadySubscriber
	}
	// Добавление нового подписчика
	currentSubscribers = append(currentSubscribers, subscriberEmail)

	// Обновление поля subscribers
	updateQuery := "UPDATE public.users SET subscribers = $1 WHERE email = $2"
	_, err = r.db.Exec(context.Background(), updateQuery, currentSubscribers, userEmail)
	if err != nil {
		return fmt.Errorf("failed to update subscribers: %v", err)
	}
	return nil
}

func (r *UserRepo) FindBirthdayPeoples(day int, month int) ([]models.User, error) {
	var birthdayPeoples []models.User

	query := `
		SELECT id, email, birthday, subscribers
		FROM public.users
		WHERE EXTRACT(DAY FROM birthday) = $1
		AND EXTRACT(MONTH FROM birthday) = $2;
	`

	rows, err := r.db.Query(context.Background(), query, day, month)
	defer rows.Close()

	if err != nil {
		return birthdayPeoples, fmt.Errorf("failed to get current birthday people: %v", err)
	}

	for rows.Next() {
		var user models.User
		var birthday time.Time
		err = rows.Scan(&user.ID, &user.Email, &birthday, &user.Subscribers)
		if err != nil {
			return birthdayPeoples, fmt.Errorf("failed to scan row: %v", err)
		}
		user.Birthday = birthday.Format(time.DateOnly)
		birthdayPeoples = append(birthdayPeoples, user)
	}

	if err = rows.Err(); err != nil {
		return birthdayPeoples, fmt.Errorf("error iterating rows: %v", err)
	}

	return birthdayPeoples, nil
}
