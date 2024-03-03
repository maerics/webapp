package web

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"webapp/db"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

const (
	PublicAssetsDirname = "public"
)

//go:embed public/*
var embedFS embed.FS

func NewServer(config Config, database *db.DB) (*Server, error) {
	gin.SetMode(config.Mode)
	engine := gin.New()
	if gin.Mode() == gin.ReleaseMode {
		engine.Use(jsonLogger())
	} else {
		engine.Use(gin.Logger())
	}

	// Initialize the cookie session.
	store := cookie.NewStore(config.CookieEncryptionKeys...)
	engine.Use(sessions.Sessions(SessionCookieName, store))

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
	log.Printf("starting web server in %q mode", s.Config.Mode)
	log.Debugf("using config %v", util.MustJson(s.Config))

	return s.Engine.Run()
}

func (s *Server) ServeStaticAssets() gin.HandlerFunc {
	var static http.Handler
	if gin.Mode() == gin.ReleaseMode {
		s.FS = http.FS(log.Must1(fs.Sub(embedFS, PublicAssetsDirname)))
	} else {
		localDirname := "web/" + PublicAssetsDirname
		log.Printf(`WARNING: frontend development mode, serving static assets from "./%s"`, localDirname)
		s.FS = http.FS(os.DirFS(localDirname))
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

// TODO: convert to regular middleware
func (s *Server) mustServeHTML(c *gin.Context, status int, filename string) {
	s.mustServeStatic(c, status, filename, ContentTypeTextHTML)
}

// TODO: convert to regular middleware
func (s *Server) mustServeStatic(c *gin.Context, status int, filename, contentType string) {
	f, err := s.FS.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	bs, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	c.Data(status, contentType, bs)
}

func jsonLogger() gin.HandlerFunc {
	type logMessage struct {
		Timestamp     string `json:"timestamp"`
		Status        int    `json:"status"`
		Method        string `json:"method"`
		Path          string `json:"path"`
		Error         string `json:"error,omitempty"`
		Latency       string `json:"latency"`
		LatencyMicros int64  `json:"latency_us"`
		ClientIP      string `json:"client_ip"`
	}

	return gin.LoggerWithFormatter(
		func(params gin.LogFormatterParams) string {
			return util.MustJson(logMessage{
				Timestamp:     params.TimeStamp.Format(log.TimestampFormat),
				Status:        params.StatusCode,
				Method:        params.Method,
				Path:          params.Path,
				Error:         params.ErrorMessage,
				Latency:       params.Latency.String(),
				LatencyMicros: params.Latency.Microseconds(),
				ClientIP:      params.ClientIP,
			}) + "\n"
		},
	)
}
func must1[T any](t T, err error) T {
	log.Must(err)
	return t
}
