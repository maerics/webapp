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
		if err := c.BindJSON(&newUser); err != nil {
			c.AbortWithError(400, err)
			return
		}

		var user models.User
		if err := s.DB.Get(&user, "INSERT INTO users (name) VALUES($1) RETURNING *", newUser.Name); err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, user)
	}
}

func (s *Server) UpdateUser() gin.HandlerFunc {
	type UpdateUserDTO struct {
		Name string `json:"name"`
	}

	return func(c *gin.Context) {
		var updateUser UpdateUserDTO
		// webMust(400, c.BindJSON(...))
		if err := c.BindJSON(&updateUser); err != nil {
			c.AbortWithError(400, err)
			return
		}

		var user models.User
		query := "UPDATE users SET name=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1 RETURNING *"
		if err := s.DB.Get(&user, query, c.Param("id"), updateUser.Name); err != nil {
			// webMust(500, s.DB.Get(...)
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, user)
	}
}

func (s *Server) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// id := webMust1(404, strconv.Atoi(...))
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithError(404, err)
			return
		}
		_, err = s.DB.Exec("DELETE FROM users WHERE id=$1", id)
		// webMust(500, err)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Status(204)
	}
}
