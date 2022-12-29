package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maerics/golog"
)

type WebErr struct {
	Context *gin.Context
	Status  int
	Err     error
}

func (s *Server) MustMiddleware() gin.HandlerFunc {
	s.Use(gin.Recovery())

	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		if webErr, ok := recovered.(WebErr); ok {
			golog.Debugf("recovered -> (%v,%q)", webErr.Status, statusMessage(webErr.Status))
			c.JSON(webErr.Status, gin.H{"error": statusMessage(webErr.Status)})
			if webErr.Status/100 == 5 {
				golog.Errorf("%v\n%v", webErr.Err, string(stack(5)))
			}
			return
		}

		// Fallback with 500 internal server error.
		c.AbortWithError(500, fmt.Errorf("%v", recovered))
		golog.Errorf("panic recovered: %v\n%v", recovered, string(stack(4)))
	})
}

func statusMessage(code int) string {
	message := strings.ToLower(http.StatusText(code))
	if message == "" {
		message = "unknown"
	}
	return message
}

func webMust(c *gin.Context, status int, err error) {
	if err != nil {
		panic(WebErr{c, status, err})
	}
}

var _ = webMust

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
