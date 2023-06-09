package internal

type User struct {
	Id       string  `db:"id" json:"id,omitempty"`
	Name     *string `db:"name" json:"name,omitempty"`
	Mail     string  `db:"mail" json:"mail" validate:"required,excludes= "`
	Password string  `db:"password" json:"password" validate:"required,excludes= "`
}
