package web

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"webapp/db"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	log "github.com/maerics/golog"
	util "github.com/maerics/goutil"
)

type Server struct {
	*gin.Engine
	Config Config
	DB     *db.DB
	FS     http.FileSystem
}

func NewServer(config Config, database *db.DB) (*Server, error) {
	engine := gin.New()
	if gin.Mode() != gin.ReleaseMode {
		engine.Use(gin.Logger())
	}

	server := &Server{
		Engine: engine,
		Config: config,
		DB:     database,
	}

	engine.Use(server.MustMiddleware())
	engine.NoRoute(server.ServeStaticAssets())

	server.GET("/_status", server.Status())
	server.ApplyRoutes()

	return server, nil
}

func (s *Server) Run() error {
	log.Printf("starting web server in %q environment", s.Config.Environment)
	log.Debugf("using config %v", util.MustJson(s.Config))
	return s.Engine.Run()
}

func (s *Server) ServeStaticAssets() gin.HandlerFunc {
	var static http.Handler
	if gin.Mode() == gin.ReleaseMode {
		s.FS = http.FS(s.Config.PublicAssets)
	} else {
		log.Print(`WARNING: frontend development mode, serving static assets from "./public"`)
		s.FS = http.FS(os.DirFS("public"))
	}
	static = http.FileServer(s.FS)
	return func(c *gin.Context) {
		f, err := s.FS.Open(path.Clean(c.Request.URL.Path))
		if errors.Is(err, fs.ErrNotExist) {
			s.notFound(c)
			return
		}
		defer f.Close()
		static.ServeHTTP(c.Writer, c.Request)
	}
}

func (s *Server) notFound(c *gin.Context) {
	panic(WebErr{c, 404, nil})
}

func (s *Server) mustServeStatic(c *gin.Context, status int, filename string) {
	f, err := s.FS.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	bs, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	mtype := mimetype.Detect(bs)
	c.Data(status, mtype.String(), bs)
}
