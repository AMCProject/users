package managers

import (
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"users/internal"
	"users/internal/models"
	"users/internal/repositories"
	"users/pkg/database"
)

type IUserManager interface {
	Login(userLogin models.User) (*models.User, error)
	GetUser(id string) (*models.User, error)
	UpdateUser(id string, userPut models.User) (*models.User, error)
	CreateUser(userPost models.User) (*models.User, error)
	DeleteUser(id string) error
}
type UserManager struct {
	db       *repositories.SQLiteUserRepository
	validate *validator.Validate
}

func NewUserManager(db database.Database) *UserManager {
	return &UserManager{
		db:       repositories.NewSQLiteUserRepository(&db),
		validate: validator.New(),
	}
}

func (u *UserManager) Login(userLogin models.User) (*models.User, error) {

	user, err := u.db.GetUserByMail(userLogin.Mail)
	if err != nil {
		return nil, err
	}

	correct := checkPasswordHash(userLogin.Password, user.Password)
	if !correct {
		return &models.User{}, internal.ErrWrongPassword
	}
	return user, nil
}

func (u *UserManager) GetUser(id string) (*models.User, error) {
	return u.db.GetUser(id)
}

func (u *UserManager) UpdateUser(id string, userUpdate models.User) (*models.User, error) {
	var err error
	if _, err = u.db.GetUser(id); err != nil {
		return nil, err
	}

	if err = u.validate.Struct(userUpdate); err != nil {
		return nil, internal.ErrWrongBody
	}

	if userUpdate.Name == nil {
		userUpdate.Name = &strings.Split(userUpdate.Mail, "@")[0]
	}

	userUpdate.Password, err = HashPassword(userUpdate.Password)
	if err != nil {
		return &models.User{}, internal.ErrHashingPassword
	}

	if err = u.db.UpdateUser(id, &userUpdate); err != nil {
		return nil, err
	}

	return u.db.GetUserByMail(userUpdate.Mail)
}

func (u *UserManager) CreateUser(userCreate models.User) (*models.User, error) {

	err := u.validate.Struct(userCreate)
	if err != nil {
		return nil, internal.ErrWrongBody
	}

	user, err := u.db.GetUserByMail(userCreate.Mail)
	if user != nil {
		return nil, internal.ErrUserAlreadyExists
	}

	if userCreate.Name == nil {
		userCreate.Name = &strings.Split(userCreate.Mail, "@")[0]
	}

	userCreate.Password, err = HashPassword(userCreate.Password)
	if err != nil {
		return nil, internal.ErrHashingPassword
	}

	if err = u.db.CreateUser(&userCreate); err != nil {
		return nil, internal.ErrSomethingWentWrong
	}

	return u.db.GetUserByMail(userCreate.Mail)

}

func (u *UserManager) DeleteUser(id string) error {
	if _, err := u.db.GetUser(id); err != nil {
		return err
	}
	return u.db.DeleteUser(id)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 7)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
