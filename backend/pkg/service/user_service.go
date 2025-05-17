package service

import (
	"coffee-loyalty-system/pkg/storage"
	"coffee-loyalty-system/pkg/storage/entities"
	"context"
)

type UserService struct {
	db storage.Database
}

func NewUserService(db storage.Database) *UserService {
	return &UserService{db: db}
}

func (s *UserService) ListUsers(ctx context.Context) ([]entities.User, error) {
	rows, err := s.db.Pool().Query(ctx, "SELECT id, email, password FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
