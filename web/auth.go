package web

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"webapp/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/maerics/golog"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "s"
	SessionUserId     = "uid"
)

func unauthorized(ctx *gin.Context) {
	webMust(ctx, 401, fmt.Errorf("Unauthorized"))
}

func (s *Server) loggedInUser(ctx *gin.Context) *models.User {
	var user models.User
	session := sessions.Default(ctx)
	switch id := session.Get(SessionUserId).(type) {
	case int:
		webMust(ctx, 500, s.DB.Get(&user, "SELECT * FROM users WHERE id=$1", id))
	}
	return nil
}

func isLoggedIn(ctx *gin.Context) bool {
	session := sessions.Default(ctx)
	switch session.Get(SessionUserId).(type) {
	case int:
		return true
	default:
		return false
	}
}

func (s *Server) LoggedInUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !isLoggedIn(ctx) {
			unauthorized(ctx)
		}
	}
}

func (s *Server) Login() gin.HandlerFunc {
	type Credentials struct{ Email, Password string }

	return func(ctx *gin.Context) {
		// Redirect home if they're already logged in.
		session := sessions.Default(ctx)
		if isLoggedIn(ctx) {
			ctx.Redirect(http.StatusFound, "/")
			return
		}

		// Parse the given credentials.
		creds := Credentials{
			Email:    ctx.PostForm("email"),
			Password: ctx.PostForm("password"),
		}
		log.Printf("OK: creds=%#v", creds)
		if isEmpty(creds.Email) || isEmpty(creds.Password) {
			unauthorized(ctx)
		}

		// Find the identified user.
		var user models.User
		query := "SELECT * FROM users WHERE email=$1"
		err := s.DB.Get(&user, query, creds.Email)
		if errors.Is(err, sql.ErrNoRows) || isEmpty(user.Password) {
			unauthorized(ctx)
		}

		// See if the login is correct.
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			unauthorized(ctx)
		}

		// Issue the login session.
		log.Printf("OK: logged in")
		session.Set(SessionUserId, user.Id)
		session.Save()
		ctx.Redirect(http.StatusFound, "/")
	}
}

func (s *Server) Logout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.SetCookie(SessionCookieName, "", -1, "/", "", true, true)
		ctx.Redirect(http.StatusFound, "/")
	}
}

func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
