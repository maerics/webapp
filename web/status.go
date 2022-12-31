package web

import (
	"github.com/gin-gonic/gin"
	util "github.com/maerics/goutil"
)

type StatusDTO struct {
	Status string      `json:"status"`
	Env    string      `json:"env"`
	Build  BuildInfo   `json:"build"`
	Net    NetworkInfo `json:"network"`
	HTTP   HTTPInfo    `json:"http"`
}

func (s *Server) Status() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := map[string]string{}
		for k := range c.Request.Header {
			headers[k] = c.Request.Header.Get(k)
		}

		c.Data(200,
			"text/json; charset=utf-8",
			[]byte(util.MustJson(&StatusDTO{
				Status: "ok",
				Env:    s.Config.Environment,
				Build:  s.Config.Build,
				Net:    getNetworkInfo(c),
				HTTP: HTTPInfo{
					Host:    c.Request.Host,
					Method:  c.Request.Method,
					URL:     c.Request.URL.String(),
					Headers: headers,
				},
			}, true)))
	}
}

type NetworkInfo struct {
	ClientIP      *string `json:"client_ip"`
	OutboundIP    *string `json:"outbound_ip,omitempty"`
	OutboundDNSIP *string `json:"outbound_dns_ip,omitempty"`
}

type HTTPInfo struct {
	Host    string            `json:"host"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func getNetworkInfo(c *gin.Context) NetworkInfo {
	var clientIP string
	if c != nil {
		clientIP = c.ClientIP()
	}

	return NetworkInfo{
		ClientIP: &clientIP,
	}
}
