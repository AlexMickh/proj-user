package models

type User struct {
	ID              string   `redis:"-"`
	Email           string   `redis:"-"`
	Name            string   `redis:"name"`
	Password        string   `redis:"password"`
	About           string   `redis:"about"`
	Skills          []string `redis:"-"`
	AvatarUrl       string   `redis:"avatar_url"`
	IsEmailVerified bool     `redis:"is_email_verified"`
}
