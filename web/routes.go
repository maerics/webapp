package web

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
)

func (s *Server) ApplyRoutes() {
	// Simple examples.
	s.GET("/hello", hello)
	s.GET("/panic", doPanic)
	s.POST("/query", s.dbQuery)

	// Cookie based login.
	s.GET("/login", func(ctx *gin.Context) { s.mustServeHTML(ctx, 200, "login.html") })
	s.POST("/login", s.Login())
	s.GET("/login/user", s.LoginAuth(), func(ctx *gin.Context) { ctx.JSON(200, s.loggedInUser(ctx)) })
	s.GET("/logout", s.Logout())

	// API group example with basic auth.
	accounts := gin.Accounts{"admin": "secret"}
	apiv1 := s.Group("/api/v1", gin.BasicAuth(accounts))
	{
		apiv1.GET("/users", s.ListUsers())
		apiv1.PUT("/users", s.CreateUser())
		apiv1.GET("/users/:id", s.GetUser())
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

// Stream SQL query results as JSON lines.
func (s *Server) dbQuery(c *gin.Context) {
	bs, err := io.ReadAll(c.Request.Body)
	webMust(c, 500, err)
	query := string(bs)
	log.Debugf("OK: query=%q", query)

	rows, err := s.DB.Query(query)
	if err != nil {
		webMust(c, 400, fmt.Errorf("invalid query: %v", err))
	}

	columns, err := rows.Columns()
	webMust(c, 500, err)

	c.Header("Content-Type", "application/ljson+json")
	bufout := bufio.NewWriterSize(c.Writer, 2*1024)
	values := make([]any, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		log.Must(rows.Scan(scanArgs...))
		fmt.Fprintf(bufout, "%v\n", util.MustJson(util.OrderedJsonObj{
			Keys:   columns,
			Values: values,
			Nulls:  true,
		}))
	}
	log.Must(bufout.Flush())
}
