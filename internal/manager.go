package internal

import (
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"users/pkg/database"
)

type IUserManager interface {
	Login(userLogin User) (*User, error)
	GetUser(id string) (*User, error)
	UpdateUser(id string, userPut User) (*User, error)
	CreateUser(userPost User) (*User, error)
	DeleteUser(id string) error
}
type UserManager struct {
	db       *SQLiteUserRepository
	validate *validator.Validate
}

func NewUserManager(db database.Database) *UserManager {
	return &UserManager{
		db:       NewSQLiteUserRepository(&db),
		validate: validator.New(),
	}
}

func (u *UserManager) Login(userLogin User) (*User, error) {

	user, err := u.db.GetUserByMail(userLogin.Mail)
	if err != nil {
		return nil, err
	}

	correct := checkPasswordHash(userLogin.Password, user.Password)
	if !correct {
		return &User{}, ErrWrongPassword
	}
	return user, nil
}

func (u *UserManager) GetUser(id string) (*User, error) {
	return u.db.GetUser(id)
}

func (u *UserManager) UpdateUser(id string, userUpdate User) (*User, error) {
	var err error
	if _, err = u.db.GetUser(id); err != nil {
		return nil, err
	}

	if err = u.validate.Struct(userUpdate); err != nil {
		return nil, ErrWrongBody
	}

	if userUpdate.Name == nil {
		userUpdate.Name = &strings.Split(userUpdate.Mail, "@")[0]
	}

	userUpdate.Password, err = hashPassword(userUpdate.Password)
	if err != nil {
		return &User{}, ErrHashingPassword
	}

	if err = u.db.UpdateUser(id, &userUpdate); err != nil {
		return nil, err
	}

	return u.db.GetUserByMail(userUpdate.Mail)
}

func (u *UserManager) CreateUser(userCreate User) (*User, error) {

	err := u.validate.Struct(userCreate)
	if err != nil {
		return nil, ErrWrongBody
	}

	user, err := u.db.GetUserByMail(userCreate.Mail)
	if user != nil {
		return nil, ErrUserAlreadyExists
	}

	if userCreate.Name == nil {
		userCreate.Name = &strings.Split(userCreate.Mail, "@")[0]
	}

	userCreate.Password, err = hashPassword(userCreate.Password)
	if err != nil {
		return nil, ErrHashingPassword
	}

	if err = u.db.CreateUser(&userCreate); err != nil {
		return nil, ErrSomethingWentWrong
	}

	return u.db.GetUserByMail(userCreate.Mail)

}

func (u *UserManager) DeleteUser(id string) error {
	if _, err := u.db.GetUser(id); err != nil {
		return err
	}
	return u.db.DeleteUser(id)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
