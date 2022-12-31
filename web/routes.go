package web

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maerics/golog"
	"github.com/maerics/goutil"
)

func (s *Server) ApplyRoutes() {
	// Simple examples.
	s.GET("/hello", hello)
	s.GET("/panic", doPanic)
	s.GET("/query", s.dbQuery)

	// API group example with basic auth.
	accounts := gin.Accounts{"admin": "secret"}
	apiv1 := s.Group("/api/v1", gin.BasicAuth(accounts))
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

// Stream SQL query results as JSON lines.
func (s *Server) dbQuery(c *gin.Context) {
	rawquery := c.Request.URL.RawQuery
	if rawquery == "" {
		webMust(c, 400, fmt.Errorf(`missing query string as SQL query`))
	}
	query, err := url.QueryUnescape(rawquery)
	webMust(c, 400, err)
	golog.Debugf("OK: query=%q", query)

	rows, err := s.DB.Query(query)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid query: %v", err)})
		return
	}

	columns, err := rows.Columns()
	webMust(c, 500, err)

	c.Header("Content-Type", "application/ljson+json")
	bufout := bufio.NewWriterSize(c.Writer, 4*1024)
	values := make([]any, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		golog.Must(rows.Scan(scanArgs...))
		fmt.Fprintf(bufout, "%v\n", goutil.MustJson(goutil.OrderedJsonObj{
			Keys:   columns,
			Values: values,
			Nulls:  true,
		}))
	}
	golog.Must(bufout.Flush())
}
