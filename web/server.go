package web

import (
	"io/fs"
	"net/http"
	"os"
	"webapp/db"

	"github.com/gin-gonic/gin"
	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
)

type Config struct {
	Environment  string
	BaseURL      string
	PublicAssets fs.FS `json:"-"`
}

type Server struct {
	*gin.Engine
	Config Config
	DB     *db.DB
}

func NewServer(config Config, database *db.DB) (*Server, error) {
	var engine *gin.Engine
	if gin.Mode() == gin.ReleaseMode {
		engine = gin.New()
		engine.Use(gin.Recovery())
	} else {
		engine = gin.Default()
	}

	server := &Server{
		Engine: engine,
		Config: config,
		DB:     database,
	}

	server.ApplyRoutes()
	engine.NoRoute(server.ServeStaticAssets())

	return server, nil
}

func (s *Server) Run() error {
	log.Debugf("starting web server with config %v", util.MustJson(s.Config))
	return s.Engine.Run()
}

// TODO: GET /_status
// TODO: notFound(c)
// TODO: internalError(c, err)

func (s *Server) ServeStaticAssets() gin.HandlerFunc {
	var static http.Handler
	if gin.Mode() == gin.ReleaseMode {
		static = http.FileServer(http.FS(s.Config.PublicAssets))
	} else {
		log.Print(`WARNING: frontend development mode, serving static assets from "./public"`)
		static = http.FileServer(http.FS(os.DirFS("public")))
	}
	return func(c *gin.Context) {
		static.ServeHTTP(c.Writer, c.Request)
	}
}
