package sniffer

import (
	"github.com/Azmus-101/azmus-agent/internal/sniffer/locate"
	"github.com/Azmus-101/azmus-agent/internal/sniffer/providers/docker"
)

// DefaultPlugins returns every plugin that ships with the agent.
func DefaultPlugins() *Plugins {
	p := NewPlugins()
	addDocker(p)
	// addNode(p)
	return p
}

func addDocker(p *Plugins) {
	p.AddLocator(locate.DockerfileLocator{})
	p.AddExtractor(docker.DockerfileExtractor{})

	p.AddLocator(locate.Compose{})
	p.AddExtractor(docker.ComposeExtractor{})
}
