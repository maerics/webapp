package web

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildInfo(t *testing.T) {
	bi := getBuildInfo()
	assert.Equal(t, runtime.Version(), bi.GoVersion)
}
