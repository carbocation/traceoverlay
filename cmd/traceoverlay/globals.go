package main

import (
	"sync"

	"github.com/carbocation/genomisc/overlay"
	"github.com/jmoiron/sqlx"
)

type Global struct {
	log logger
	db  *sqlx.DB

	Site      string
	Company   string
	Email     string
	SnailMail string

	Project      string
	ManifestPath string
	PreviewAlpha int

	m        sync.RWMutex
	manifest []Manifest

	Config overlay.JSONConfig
}

func (g *Global) Manifest() []Manifest {
	g.m.RLock()
	defer g.m.RUnlock()

	return g.manifest
}

type logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
