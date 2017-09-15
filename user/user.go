package user

import (
	"database/sql"
	"errors"
	"net/mail"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id             string
	Username       string
	Email          string
	passwordHash   []byte
	twoFactorToken string
	CreatedAt      time.Time
	Deleted        bool
	store          UserStore
}

func (u *User) SetPassword(password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.passwordHash = passwordHash

	return u.store.SetUserPassword(u.Id, passwordHash)
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.passwordHash, []byte(password))
	return err == nil
}

type UserStore interface {
	GetUser(string, string) *User
	CreateUser(*User) error
	SetUserPassword(string, []byte) error
}

type UserStoreFile struct {
	db *sql.DB
}

func (s *UserStoreFile) GetUser(key string, value string) *User {
	if key != "Id" && key != "Username" && key != "Email" {
		return nil
	}

	if key == "Username" {
		value, _ = cannonicalUsername(value)
	} else if key == "email" {
		value, _ = cannonicalEmail(value)
	}

	rows, err := s.db.Query(`SELECT
		Id,
		Username,
		Email,
		PasswordHash,
		TwoFactorToken,
		CreatedAt,
		Deleted FROM users WHERE `+key+` = ? COLLATE NOCASE;`, value)
	if err != nil {
		return nil
	}

	user := &User{store: s}
	for rows.Next() {
		err = rows.Scan(&user.Id, &user.Username, &user.Email, &user.passwordHash, &user.twoFactorToken, &user.CreatedAt, &user.Deleted)
		if err != nil {
			return nil
		}
	}

	if len(user.Id) == 0 {
		return nil
	}

	return user
}

func (s *UserStoreFile) CreateUser(user *User) error {
	stmt, err := s.db.Prepare(`INSERT INTO users (
		Id,
		Username,
		Email,
		PasswordHash,
		TwoFactorToken,
		CreatedAt,
		Deleted
	)  values(?,?,?,?,?,?,?);`)

	if err != nil {
		return err
	}

	if len(user.Id) == 0 {
		user.Id = uuid.NewV4().String()
	}

	if user.Email, err = cannonicalEmail(user.Email); err != nil {
		return err
	}

	if user.Username, err = cannonicalUsername(user.Username); err != nil {
		return err
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	if user.store == nil {
		user.store = s
	}

	_, err = stmt.Exec(user.Id, user.Username, user.Email, user.passwordHash, user.twoFactorToken, user.CreatedAt, false)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStoreFile) SetUserPassword(id string, passwordHash []byte) error {
	stmt, err := s.db.Prepare(`UPDATE users SET PasswordHash = ? WHERE Id = ?;`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(passwordHash, id)
	if err != nil {
		return err
	}

	return nil
}

func NewFileStore(filename string) (UserStore, error) {
	if stat, err := os.Stat(filename); os.IsNotExist(err) || stat.Size() == 0 {
		return CreateFileStore(filename)
	}

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	return &UserStoreFile{db: db}, nil
}

func CreateFileStore(filename string) (UserStore, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(`CREATE TABLE users (
		Id text NOT NULL PRIMARY KEY COLLATE NOCASE,
		Username text NOT NULL UNIQUE COLLATE NOCASE,
		Email text NOT NULL UNIQUE COLLATE NOCASE,
		PasswordHash text,
		TwoFactorToken text,
		CreatedAt timestamp,
		Deleted boolean
	);`)

	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &UserStoreFile{db: db}, nil
}

func cannonicalEmail(email string) (string, error) {
	if cEmail, err := mail.ParseAddress(email); err != nil {
		return "", errors.New("Unparsable email address: " + email)
	} else {
		return cEmail.Address, nil
	}
}

func cannonicalUsername(username string) (string, error) {
	return strings.TrimSpace(username), nil
}
