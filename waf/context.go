package waf

import (
	"log"
	"time"

	"github.com/casualjim/go-swagger/spec"
)

type Context struct {
	Doc              *spec.Document
	Listen, Backend  string
	LogOnly, Verbose bool
}

/* Reanalyze the spec */
func (c *Context) Refresh() {
	for {
		time.Sleep(5 * time.Minute)
		c.Doc = c.Doc.Reload()
	}
}

func (c *Context) load(path string) {
	doc, err := spec.JSONSpec(path)
	if err != nil {
		log.Println("Unable to log Swagger file!")
		panic(err)
	}
	c.Doc = doc
}

func InitContext(listen, backend, swagger string, logOnly, verbose bool) *Context {
	ctx := &Context{
		Listen:  listen,
		Backend: backend,
		LogOnly: logOnly,
		Verbose: verbose,
	}
	ctx.load(swagger)
	go ctx.Refresh()
	return ctx
}
