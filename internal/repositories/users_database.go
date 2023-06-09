package repositories

import (
	"github.com/labstack/gommon/log"
	"github.com/oklog/ulid/v2"
	"math/rand"
	"time"
	"users/internal"
	"users/internal/models"
	"users/pkg/database"
)

const (
	getUser     = "SELECT * FROM users WHERE id = ?"
	getUserMail = "SELECT * FROM users WHERE mail = ?"
	createUser  = "INSERT INTO users(id,name,mail,password) VALUES (?,?,?,?)"
	updateUser  = "UPDATE users SET name = ?, mail = ?, password = ? WHERE id = ?"
	deleteUser  = "DELETE FROM users WHERE id = ?"
)

type SQLiteUserRepository struct {
	db *database.Database
}

type UserRepository interface {
	GetUser(id string) (*models.User, error)
	GetUserByMail(mail string) (*models.User, error)
	UpdateUser(user *models.User) error
	CreateUser(user *models.User) error
	DeleteUser(id string) error
}

func NewSQLiteUserRepository(db *database.Database) *SQLiteUserRepository {
	return &SQLiteUserRepository{
		db: db,
	}
}

func (r *SQLiteUserRepository) GetUser(id string) (user *models.User, err error) {
	var usersAux []models.User

	if err = r.db.Conn.Select(&usersAux, getUser, id); err != nil {
		log.Error(err)
		return user, internal.ErrSomethingWentWrong
	}
	if len(usersAux) == 0 {
		return user, internal.ErrUserNotFound
	}

	return &usersAux[0], nil
}

func (r *SQLiteUserRepository) GetUserByMail(mail string) (user *models.User, err error) {
	var usersAux []models.User

	if err = r.db.Conn.Select(&usersAux, getUserMail, mail); err != nil {
		log.Error(err)
		return user, internal.ErrSomethingWentWrong
	}

	if len(usersAux) == 0 {
		return user, internal.ErrUserNotFound
	}
	return &usersAux[0], nil
}

func (r *SQLiteUserRepository) UpdateUser(id string, user *models.User) (err error) {
	if _, err = r.db.Conn.Exec(updateUser, user.Name, user.Mail, user.Password, id); err != nil {
		log.Error(err)
		return internal.ErrSomethingWentWrong
	}
	return
}

func (r *SQLiteUserRepository) CreateUser(user *models.User) (err error) {

	id, _ := ulid.New(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0))
	if _, err = r.db.Conn.Exec(createUser, id.String(), user.Name, user.Mail, user.Password); err != nil {
		log.Error(err)
		return internal.ErrSomethingWentWrong
	}
	return
}

func (r *SQLiteUserRepository) DeleteUser(id string) (err error) {

	if _, err = r.db.Conn.Exec(deleteUser, id); err != nil {
		log.Error(err)
		return internal.ErrSomethingWentWrong
	}
	return
}
