package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// ReverseProxyHandler returns a Gin HandlerFunc that proxies incoming requests to targetURL.
func ReverseProxyHandler(targetURL string) gin.HandlerFunc {
	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid target microservice URL: %s, error: %v", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	// Customize HTTP request director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Maintain downstream hostname matching
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = url.Host
	}

	// Graceful error responder in case downstream microservice is offline
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error": "Bad Gateway", "message": "Downstream microservice is currently unreachable or starting up."}`))
	}

	// Clean up duplicate CORS headers returned from downstream microservices
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Expose-Headers")
		return nil
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
