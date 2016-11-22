package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gopkg.in/authboss.v0"
)

var nextUserID int

type User struct {
	ID   int    `db:"ID"`
	Key  string `db:"Key"`
	Name string `db:"Name"`

	// Auth
	Email    string `db:"Email"`
	Password string `db:"Password"`

	// OAuth2
	Oauth2Uid      string     `db:"Oauth2Uid"`
	Oauth2Provider string     `db:"Oauth2Provider"`
	Oauth2Token    string     `db:"Oauth2Token"`
	Oauth2Refresh  string     `db:"Oauth2Refresh"`
	Oauth2Expiry   *time.Time `db:"Oauth2Expiry"`

	// Confirm
	ConfirmToken string `db:"ConfirmToken"`
	Confirmed    bool   `db:"Confirmed"`

	// Lock
	AttemptNumber int64      `db:"AttemptNumber"`
	AttemptTime   *time.Time `db:"AttemptTime"`
	Locked        *time.Time `db:"Locked"`

	// Recover
	RecoverToken       string     `db:"RecoverToken"`
	RecoverTokenExpiry *time.Time `db:"RecoverTokenExpiry"`

	// Remember is in another table
}

type DBStorer struct {
	db     *sqlx.DB
	Tokens map[string][]string
}

func NewDBStorer(databaseDriverName, databaseDataSourceName string) *DBStorer {
	db, err := sqlx.Connect(databaseDriverName, databaseDataSourceName)
	if err != nil {
		panic(err)
	}

	return &DBStorer{
		db:     db,
		Tokens: make(map[string][]string),
	}
}

func (c *DBStorer) Create(key string, attr authboss.Attributes) error {
	var user User
	if err := attr.Bind(&user, true); err != nil {
		return err
	}
	user.Key = key

	// ID
	// Key
	// Name
	// Email
	// Password
	// Oauth2Uid
	// Oauth2Provider
	// Oauth2Token
	// Oauth2Refresh
	// Oauth2Expiry
	// ConfirmToken
	// Confirmed
	// AttemptNumber
	// AttemptTime
	// Locked
	// RecoverToken
	// RecoverTokenExpiry
	//  (ID,Key,Name,Email,Password,Oauth2Uid,Oauth2Provider,Oauth2Token,Oauth2Refresh,Oauth2Expiry,ConfirmToken,Confirmed,AttemptNumber,AttemptTime,Locked,RecoverToken,RecoverTokenExpiry)
	if _, err := c.db.NamedExec("INSERT INTO users VALUES (:ID,:Key,:Name,:Email,:Password,:Oauth2Uid,:Oauth2Provider,:Oauth2Token,:Oauth2Refresh,:Oauth2Expiry,:ConfirmToken,:Confirmed,:AttemptNumber,:AttemptTime,:Locked,:RecoverToken,:RecoverTokenExpiry)", &user); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (c *DBStorer) Put(key string, attr authboss.Attributes) error {
	return c.Create(key, attr)
}

func (c *DBStorer) Get(key string) (result interface{}, err error) {
	var user User

	if err := c.db.Get(&user, "SELECT * FROM users WHERE Key = ?", key); err != nil {
		return nil, authboss.ErrUserNotFound
	}

	return &user, nil
}

func (c *DBStorer) PutOAuth(uid, provider string, attr authboss.Attributes) error {
	return c.Create(uid+provider, attr)
}

func (c *DBStorer) GetOAuth(uid, provider string) (result interface{}, err error) {
	return c.Get(uid + provider)
}

func (c *DBStorer) AddToken(key, token string) error {
	c.Tokens[key] = append(c.Tokens[key], token)
	fmt.Println("AddToken")
	spew.Dump(c.Tokens)
	return nil
}

func (c *DBStorer) DelTokens(key string) error {
	delete(c.Tokens, key)
	fmt.Println("DelTokens")
	spew.Dump(c.Tokens)
	return nil
}

func (c *DBStorer) UseToken(givenKey, token string) error {
	toks, ok := c.Tokens[givenKey]
	if !ok {
		return authboss.ErrTokenNotFound
	}

	for i, tok := range toks {
		if tok == token {
			toks[i], toks[len(toks)-1] = toks[len(toks)-1], toks[i]
			c.Tokens[givenKey] = toks[:len(toks)-1]
			return nil
		}
	}

	return authboss.ErrTokenNotFound
}

func (c *DBStorer) ConfirmUser(tok string) (result interface{}, err error) {
	fmt.Println("==============", tok)

	var user User

	if err := c.db.Get(&user, "SELECT * FROM users WHERE ConfirmToken = ?", tok); err != nil {
		return nil, authboss.ErrUserNotFound
	}

	return &user, nil
}

func (c *DBStorer) RecoverUser(rec string) (result interface{}, err error) {
	var user User

	if err := c.db.Get(&user, "SELECT * FROM users WHERE RecoverToken = ?", rec); err != nil {
		return nil, authboss.ErrUserNotFound
	}

	return &user, nil
}

type MemStorer struct {
	Users  map[string]User
	Tokens map[string][]string
}

func NewMemStorer() *MemStorer {
	return &MemStorer{
		Users: map[string]User{
			"zeratul@heroes.com": User{
				ID:        1,
				Name:      "Zeratul",
				Password:  "$2a$10$XtW/BrS5HeYIuOCXYe8DFuInetDMdaarMUJEOg/VA/JAIDgw3l4aG", // pass = 1234
				Email:     "zeratul@heroes.com",
				Confirmed: true,
			},
		},
		Tokens: make(map[string][]string),
	}
}

func (s MemStorer) Create(key string, attr authboss.Attributes) error {
	var user User
	if err := attr.Bind(&user, true); err != nil {
		return err
	}

	user.ID = nextUserID
	nextUserID++

	s.Users[key] = user
	fmt.Println("Create")
	spew.Dump(s.Users)
	return nil
}

func (s MemStorer) Put(key string, attr authboss.Attributes) error {
	return s.Create(key, attr)
}

func (s MemStorer) Get(key string) (result interface{}, err error) {
	user, ok := s.Users[key]
	if !ok {
		return nil, authboss.ErrUserNotFound
	}

	return &user, nil
}

func (s MemStorer) PutOAuth(uid, provider string, attr authboss.Attributes) error {
	return s.Create(uid+provider, attr)
}

func (s MemStorer) GetOAuth(uid, provider string) (result interface{}, err error) {
	user, ok := s.Users[uid+provider]
	if !ok {
		return nil, authboss.ErrUserNotFound
	}

	return &user, nil
}

func (s MemStorer) AddToken(key, token string) error {
	s.Tokens[key] = append(s.Tokens[key], token)
	fmt.Println("AddToken")
	spew.Dump(s.Tokens)
	return nil
}

func (s MemStorer) DelTokens(key string) error {
	delete(s.Tokens, key)
	fmt.Println("DelTokens")
	spew.Dump(s.Tokens)
	return nil
}

func (s MemStorer) UseToken(givenKey, token string) error {
	toks, ok := s.Tokens[givenKey]
	if !ok {
		return authboss.ErrTokenNotFound
	}

	for i, tok := range toks {
		if tok == token {
			toks[i], toks[len(toks)-1] = toks[len(toks)-1], toks[i]
			s.Tokens[givenKey] = toks[:len(toks)-1]
			return nil
		}
	}

	return authboss.ErrTokenNotFound
}

func (s MemStorer) ConfirmUser(tok string) (result interface{}, err error) {
	fmt.Println("==============", tok)

	for _, u := range s.Users {
		if u.ConfirmToken == tok {
			return &u, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}

func (s MemStorer) RecoverUser(rec string) (result interface{}, err error) {
	for _, u := range s.Users {
		if u.RecoverToken == rec {
			return &u, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}
