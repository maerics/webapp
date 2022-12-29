package web

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) ApplyRoutes() {
	s.GET("/hello", func(c *gin.Context) { c.JSON(200, "Hello, World!") })

	apiv1 := s.Group("/api/v1")
	{
		apiv1.GET("/users", s.ListUsers())
		apiv1.PUT("/users", s.CreateUser())
		apiv1.POST("/users/:id", s.UpdateUser())
		// apiv1.DELETE("/users/:id", s.DeleteUser())
	}
}
