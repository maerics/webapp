package web

import (
	"database/sql"
	"strconv"
	"webapp/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) ListUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		users := []models.User{}
		webMust(c, 500, s.DB.Select(&users, "SELECT * FROM users ORDER BY id"))
		c.JSON(200, users)
	}
}

func (s *Server) CreateUser() gin.HandlerFunc {
	type NewUserDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c *gin.Context) {
		var newUser NewUserDTO
		webMust(c, 400, c.BindJSON(&newUser))

		var user models.User
		query := "INSERT INTO users (email,password) VALUES($1,$2) RETURNING *"
		webMust(c, 500, s.DB.Get(&user, query, newUser.Email, BCryptPassword(newUser.Password)))

		c.JSON(200, models.User{
			Id:        user.Id,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
}

func (s *Server) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		webMust(c, 404, err)

		var user models.User
		err = s.DB.Get(&user, "SELECT * FROM users WHERE id=$1", id)
		if err == sql.ErrNoRows {
			s.notFound(c)
		}
		webMust(c, 500, err)

		c.JSON(200, models.User{
			Id:        user.Id,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
}

func (s *Server) UpdateUser() gin.HandlerFunc {
	type UpdateUserDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		webMust(c, 404, err)

		var updateUser UpdateUserDTO
		webMust(c, 400, c.BindJSON(&updateUser))

		var user models.User
		query := "UPDATE users SET email=$1, password=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$3 RETURNING *"
		err = s.DB.Get(&user, query, updateUser.Email, BCryptPassword(updateUser.Password), id)
		if err == sql.ErrNoRows {
			s.notFound(c)
		}
		webMust(c, 500, err)

		c.JSON(200, user)
	}
}

func (s *Server) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		webMust(c, 404, err)

		_, err = s.DB.Exec("DELETE FROM users WHERE id=$1", id)
		webMust(c, 500, err)
		c.Status(204)
	}
}

func BCryptPassword(s string) string {
	return string(must1(bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)))
}
