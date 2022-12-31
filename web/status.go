package web

import (
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/maerics/golog"
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
		ClientIP:      &clientIP,
		OutboundIP:    getOutboundIP(),
		OutboundDNSIP: getOutboundDNSIP(),
	}
}

func getOutboundIP() *string {
	res, err := http.Get("http://checkip.amazonaws.com/")
	if err != nil {
		log.Errorf("failed to determine outbound IP: %v", err)
		return nil
	}
	if res.StatusCode != 200 {
		log.Errorf("unexpected response for outbound IP: %v", res.StatusCode)
		return nil
	}
	bs, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("error reading outbound IP: %v", err)
		return nil
	}
	ip := strings.TrimSpace(string(bs))
	return &ip
}

func getOutboundDNSIP() *string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Errorf("failed to determine outbound DNS IP: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()
	return &ip
}
