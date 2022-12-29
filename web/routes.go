package web

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (s *Server) ApplyRoutes() {
	s.GET("/hello", hello)

	apiv1 := s.Group("/api/v1")
	{
		apiv1.GET("/users", s.ListUsers())
		apiv1.PUT("/users", s.CreateUser())
		apiv1.POST("/users/:id", s.UpdateUser())
		apiv1.DELETE("/users/:id", s.DeleteUser())
	}
}

func hello(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		name = "World"
	}
	c.JSON(200, gin.H{
		"greetings": fmt.Sprintf("Hello, %v!", name),
	})
}
