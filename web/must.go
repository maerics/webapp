package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/maerics/golog"
)

const (
	ContentTypeTextHTML = "text/html; charset=utf-8"

	DefaultFilename404 = "404.html"
	DefaultFilename500 = "500.html"
)

type WebErr struct {
	Context *gin.Context
	Status  int
	Err     error
}

func webMust(c *gin.Context, status int, err error) {
	if err != nil {
		panic(WebErr{c, status, err})
	}
}

// Recovery middleware which enables using the "webMust(...)" function
// which abuses panics to avoid repetitive boilerplate error handling.
func (s *Server) MustMiddleware() gin.HandlerFunc {
	const jsonContentType = "text/json; charset=utf-8"

	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		// Handle WebErr types specially.
		if webErr, ok := recovered.(WebErr); ok {
			log.Debugf("recovered WebErr -> (%v,%q,%v)",
				webErr.Status, statusMessage(webErr.Status), webErr.Err)

			// Always print the stack trace for 5xx class errors.
			if webErr.Status/100 == 5 {
				log.Errorf("%v\n%v", webErr.Err, string(stack(5)))
			}

			filename404 := firstNonEmpty(s.Config.Filename404, DefaultFilename404)
			filename500 := firstNonEmpty(s.Config.Filename500, DefaultFilename500)
			respond404 := func() { s.mustServeHTML(c, 404, filename404) }
			respond500 := func() { s.mustServeHTML(c, webErr.Status, filename500) }
			if preferJson(c.Request.Header) {
				respond404 = func() { c.Data(404, jsonContentType, []byte(`{"error":"not found"}`)) }
				respond500 = func() { c.Data(500, jsonContentType, []byte(`{"error":"internal server error"}`)) }
			}

			// Respond with common 404, 500, or JSON responses.
			switch true {
			case webErr.Status == 404:
				respond404()
			case webErr.Status/100 == 4:
				c.JSON(webErr.Status, gin.H{"error": webErr.Err.Error()})
			case webErr.Status/100 == 5:
				c.Error(webErr.Err)
				respond500()
			default:
				c.Error(webErr.Err)
				c.JSON(webErr.Status, gin.H{"error": statusMessage(webErr.Status)})
			}
			c.Abort()
			return
		}

		// Fallback with 500 internal server error.
		if preferJson(c.Request.Header) {
			c.Data(500, jsonContentType, []byte(`{"error":"internal server error"}`))
		} else {
			s.mustServeHTML(c, 500, firstNonEmpty(s.Config.Filename500, DefaultFilename500))
		}
		c.AbortWithError(500, fmt.Errorf("%v", recovered))
		log.Errorf("panic recovered: %v\n%v", recovered, string(stack(4)))
	})
}

// Shoddy content negotation for JSON preference.
var preferJson = (func() func(http.Header) bool {
	commaSepRegex := regexp.MustCompile(`\s*,\s*`)
	htmlTypeRegex := regexp.MustCompile(`/html\b`)
	jsonTypeRegex := regexp.MustCompile(`/json\b`)
	return func(header http.Header) bool {
		// See if they prefer to accept JSON.
		acceptHeader := header.Get("Accept")
		parts := commaSepRegex.Split(acceptHeader, -1)
		for _, part := range parts {
			switch true {
			case htmlTypeRegex.MatchString(part):
				return false
			case jsonTypeRegex.MatchString(part):
				return true
			}
		}
		// See if they sent JSON.
		return jsonTypeRegex.MatchString(header.Get("Content-Type"))
	}
})()

func statusMessage(code int) string {
	message := strings.ToLower(http.StatusText(code))
	if message == "" {
		message = "unknown"
	}
	return message
}

// See https://github.com/gin-gonic/gin/blob/master/recovery.go
var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contain dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
