package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)


type ServiceProxy struct {
	TargetURL string
	Proxy     *httputil.ReverseProxy
}


func NewServiceProxy(targetURL string) (*ServiceProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL %s: %w", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	
	
	proxy.Transport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().Err(err).Str("target", targetURL).Msg("Proxy error")
		http.Error(w, "Service temporarily unavailable", http.StatusBadGateway)
	}

	
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Add gateway headers
		resp.Header.Set("X-Gateway", "SOA-Gateway")
		resp.Header.Set("X-Forwarded-By", "Custom-Go-Gateway")
		return nil
	}

	return &ServiceProxy{
		TargetURL: targetURL,
		Proxy:     proxy,
	}, nil
}


func (sp *ServiceProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	
	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("target", sp.TargetURL).
		Str("user_agent", r.UserAgent()).
		Msg("Proxying request")

	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	r = r.WithContext(ctx)

	
	sp.Proxy.ServeHTTP(w, r)
}




type ServiceRegistry struct {
	proxies map[string]*ServiceProxy
}


func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		proxies: make(map[string]*ServiceProxy),
	}
}


func (sr *ServiceRegistry) RegisterService(name, targetURL string) error {
	proxy, err := NewServiceProxy(targetURL)
	if err != nil {
		return err
	}
	
	sr.proxies[name] = proxy
	log.Info().Str("service", name).Str("target", targetURL).Msg("Service registered")
	return nil
}


func (sr *ServiceRegistry) GetService(name string) (*ServiceProxy, bool) {
	proxy, exists := sr.proxies[name]
	return proxy, exists
}


func (sr *ServiceRegistry) ListServices() []string {
	services := make([]string, 0, len(sr.proxies))
	for name := range sr.proxies {
		services = append(services, name)
	}
	return services
}


