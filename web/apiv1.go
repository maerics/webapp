package web

import (
	"strconv"
	"webapp/models"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		users := []models.User{}
		if err := s.DB.Select(&users, "SELECT * FROM users ORDER BY id"); err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, users)
	}
}

func (s *Server) CreateUser() gin.HandlerFunc {
	type NewUserDTO struct {
		Name string `json:"name"`
	}

	return func(c *gin.Context) {
		var newUser NewUserDTO
		webMust(c, 400, c.BindJSON(&newUser))

		var user models.User
		err := s.DB.Get(&user, "INSERT INTO users (name) VALUES($1) RETURNING *", newUser.Name)
		webMust(c, 500, err)

		c.JSON(200, user)
	}
}

func (s *Server) UpdateUser() gin.HandlerFunc {
	type UpdateUserDTO struct {
		Name string `json:"name"`
	}

	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		webMust(c, 404, err)

		var updateUser UpdateUserDTO
		webMust(c, 400, c.BindJSON(&updateUser))

		var user models.User
		query := "UPDATE users SET name=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2 RETURNING *"
		webMust(c, 500, s.DB.Get(&user, query, updateUser.Name, id))

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
