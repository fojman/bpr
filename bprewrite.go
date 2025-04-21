// Package rewrite contains custom BePlay rewrite logic.
package rewrite

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

// Provider config
type Provider struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	RTP     string `yaml:"rtp"`
	Super   string `yaml:"super"`
}

// Config holds the plugin configuration.
type Config struct {
	Providers []Provider `yaml:"providers"`
}

// CreateConfig creates a new Config with default values.
func CreateConfig() *Config {
	return &Config{
		Providers: []Provider{},
	}
}

// Rewrite holds the plugin state.
type Rewrite struct {
	next   http.Handler
	name   string
	config *Config
}

func init() {
	log.SetOutput(os.Stdout)
}

// New creates a new DemoPlugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config == nil {
		config = CreateConfig()
	}

	return &Rewrite{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

// ServeHTTP implements the http.Handler interface.
func (p *Rewrite) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Println("this is the log mesage from plugin `rewrite middleware`")

	provName := req.Header.Get("X-Prov-Id")
	if provName == "" {
		log.Println("no X-Prov-Id in req")
		p.next.ServeHTTP(rw, req) // should we raise error here?!
		return
	}

	// Find matching provider
	var provider *Provider
	for _, p := range p.config.Providers {
		if p.Name == provName {
			provider = &p
			break
		}
	}

	if provider == nil {
		log.Printf("No provider found for X-Prov: %s - try `default'?", provName)
		p.next.ServeHTTP(rw, req)
		return
	}

	// Parse the path: expect /<provider>/<game>
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != provName {
		log.Printf("Invalid path format: %s, expected /%s/<game>", req.URL.Path, provName)
		p.next.ServeHTTP(rw, req)
		return
	}

	game := parts[1]

	// Construct new path: /dev/<game>/config.<rtp>/config.<version>
	newPath := path.Join("/dev", game, provider.RTP, provider.Version)
	log.Printf("rewrite: %s -> %s", req.URL.Path, newPath)

	//req.URL.Path = newPath
	//req.RequestURI = newPath // Update RequestURI for downstream handlers

	//log.Printf("Rewrote path from %s to %s for provider %s", req.URL.Path, newPath, provName)

	// Pass to next handler
	p.next.ServeHTTP(rw, req)
}
