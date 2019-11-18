package main

import (
	"database/sql"
	"errors"

	"./models"
	"github.com/gorilla/sessions"
)

// Context type represents context
type Context map[string]interface{}

// EmptyContext ...
func EmptyContext() Context {
	ctx := make(map[string]interface{})

	return ctx
}

func createContextFromSession(cr *sql.DB, session *sessions.Session) (Context, error) {
	ctx := EmptyContext()

	authenticated, ok := session.Values["authenticated"].(bool)

	if !ok {
		err := "Authenticated not present in session"
		return EmptyContext(), errors.New(err)
	}

	uid, ok := session.Values["uid"].(int)

	if !ok {
		err := "UserID not present in session"
		return EmptyContext(), errors.New(err)
	}

	ctx["Authenticated"] = authenticated
	ctx["User"] = models.FindUserByID(cr, uid)

	return ctx, nil
}
