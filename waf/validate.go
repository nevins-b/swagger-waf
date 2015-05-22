package waf

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/casualjim/go-swagger/spec"
)

type ValidatorResponse struct {
	check  string
	result bool
	err    *error
}

func Valid(ctx *Context, r *http.Request) bool {
	c := make(chan bool)

	checks := []func(*Context, *http.Request, chan bool){
		checkHost,
		checkPathAndMethod,
		checkContentType,
		checkContentLength,
		checkParameters,
	}

	for i := range checks {
		go checks[i](ctx, r, c)
	}

	var responses []bool
	var resp bool
	for {
		resp = <-c
		if !resp {
			return ctx.LogOnly || false
		}
		responses = append(responses, resp)
		if len(responses) == len(checks) {
			break
		}
	}
	return true
}

func cleanURI(ctx *Context, uri string) string {
	base := ctx.Doc.BasePath()
	return strings.Replace(uri, base, "", 1)
}

func getOperation(ctx *Context, r *http.Request) (*spec.Operation, bool) {
	uri := cleanURI(ctx, r.URL.RequestURI())
	method := r.Method
	return ctx.Doc.OperationFor(method, uri)
}

func getBody(r *http.Request) (*bytes.Buffer, error) {
	contentLength, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		contentLength = 0
	}
	buf := make([]byte, contentLength)
	n, err := r.Body.Read(buf)
	if err != nil {
		log.Printf("Error getting message body: %v", err)
	}
	if n != contentLength {
		return nil, errors.New("Content-Length Is not equal")
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	return bytes.NewBuffer(buf), nil
}

func checkContentType(ctx *Context, r *http.Request, c chan bool) {
	/* Content Type doesn't matter for GET */
	if r.Method == "GET" {
		c <- true
		return
	}
	op, found := getOperation(ctx, r)
	if !found {
		log.Printf("Failed to find operation")
		c <- false
		return
	}
	if len(op.Consumes) == 0 {
		log.Printf("Content Type check returned: true")
		c <- true
		return
	}
	for _, t := range op.Consumes {
		log.Printf(t)
		if t == r.Header.Get("Content-Type") {
			log.Printf("Content Type check returned: true")
			c <- true
			return
		}
	}
	log.Printf("Content Type check returned: false")
	c <- false
}

func checkContentLength(ctx *Context, r *http.Request, c chan bool) {
	_, err := getBody(r)
	if err != nil {
		c <- false
		return
	}
	c <- true
}

func checkHost(ctx *Context, r *http.Request, c chan bool) {
	match := r.Host == ctx.Doc.Host()
	if ctx.Verbose {
		log.Printf("Host Check returned: %v", match)
	}
	c <- match
}

func checkPathAndMethod(ctx *Context, r *http.Request, c chan bool) {
	_, found := getOperation(ctx, r)
	if ctx.Verbose {
		log.Printf("Path And Method Check returned: %v", found)
	}
	c <- found
}

func checkParameters(ctx *Context, r *http.Request, c chan bool) {
	op, found := getOperation(ctx, r)
	if !found {
		log.Printf("Failed to find operation")
		c <- found
		return
	}

	if r.Method == "GET" {
		for _, v := range op.Parameters {
			data := r.FormValue(v.Name)
			if data == "" && v.Required {
				log.Printf("Parameter Check return: false, required param %s unset.", v.Name)
				c <- false
				return
			}
		}
	} else {
		body, err := getBody(r)
		if err != nil {
			c <- false
			return
		}
		if len(op.Consumes) > 0 || len(op.Parameters) == 1 {
			var objmap map[string]*json.RawMessage
			err := json.Unmarshal(body.Bytes(), &objmap)
			if err != nil {
				log.Printf("Error decoding JSON: %v", err)
				c <- false
				return
			}
			for _, v := range op.Parameters {
				if v.Schema != nil {
					log.Printf("Checking Schema with %d properties", len(v.Schema.Properties))
					for name, sc := range v.Schema.Properties {
						data, ok := objmap[name]
						if !ok && stringInSlice(v.Schema.Required, name) {
							c <- false
							return
						}
						var b []byte
						err := data.UnmarshalJSON(b)
						log.Printf("Parsing %s as %s", string(b), sc.Format)
						_, err = stringToType(string(b), sc.Format)
						if err != nil {
							c <- false
							return
						}
					}
				}
			}
		} else {
			for _, v := range op.Parameters {
				data := r.FormValue(v.Name)
				if data == "" && v.Required {
					log.Printf("Parameter Check return: false, required param %s unset.", v.Name)
					c <- false
					return
				}
			}
		}
	}
	log.Printf("Parameter Check returned: true")
	c <- true
}
