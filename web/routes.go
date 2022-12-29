package web

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) ApplyRoutes() {
	s.GET("/hello", hello)
	s.GET("/panic", doPanic)

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

func doPanic(c *gin.Context) {
	message := c.Query("message")
	if message == "" {
		message = "panic"
	}

	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		webMust(c, status, errors.New(message))
	}

	if strings.TrimSpace(c.Query("raw")) != "" {
		panic(message)
	}
	panic(errors.New(message))
}
