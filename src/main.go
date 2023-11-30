package main

// this was created while experimenting with json-ld parsing. Never used, but sharing as existing documentation is not great.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/piprate/json-gold/ld"
)

// see README for details on how to get this file
const f = "./schemaorgcontext.jsonld"

// LdProcessor is a processor for json-ld
type LdProcessor interface {
	Parse(map[string]interface{}) (interface{}, error)
}

type ldProcessor struct {
	proc    *ld.JsonLdProcessor
	options *ld.JsonLdOptions
}

// NewLdProcessor returns a new LdProcessor
func NewLdProcessor(client http.Client, caching bool) LdProcessor {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	// add caching
	nl := ld.NewDefaultDocumentLoader(&client)
	cdl := ld.NewCachingDocumentLoader(nl)

	var m map[string]string

	if caching {
		// document mapping
		m = map[string]string{
			"https://schema.org/": f,
			"http://schema.org/":  f,
			"https://schema.org":  f,
			"http://schema.org":   f,
		}
	}

	// PreloadWithMapping populates the cache, in this case from local files
	if err := cdl.PreloadWithMapping(m); err != nil {
		panic(err)
	}

	// set the document loader
	options.DocumentLoader = cdl

	return &ldProcessor{
		proc:    proc,
		options: options,
	}
}

func (s *ldProcessor) Parse(i map[string]interface{}) (interface{}, error) {
	expanded, err := s.proc.Expand(i, s.options)
	if err != nil {
		return nil, err
	}

	if expanded == nil || len(expanded) != 1 {
		return nil, fmt.Errorf("unexpected number of expanded graphs: %d", len(expanded))
	}

	// ld.PrintDocument("JSON-LD expansion succeeded", expanded)
	return expanded[0], nil
}

// loggingTransport and RoundTrip are used to debug http requests
// https://www.jvt.me/posts/2023/03/11/go-debug-http/
type loggingTransport struct{}

func (s *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bytes, _ := httputil.DumpRequestOut(r, true)

	resp, err := http.DefaultTransport.RoundTrip(r)
	// err is returned after dumping the response

	respBytes, _ := httputil.DumpResponse(resp, true)
	bytes = append(bytes, respBytes...)

	fmt.Printf("%s\n", bytes)

	return resp, err
}

func main() {
	// todo: toggle debugging
	client := http.Client{
		Transport: &loggingTransport{},
	}

	recipe := `{
		"@context": "https://schema.org/",
		"@type": "Recipe",
		"name": "Overnight chocolate oats",
		"recipeIngredient": [
			"1 container Silken tofu",
			"1 cup Steal cut oats",
			"2 tbsp Coco powder",
			"2 tbsp Maple syrup",
			" ",
			"1 Banana",
			"1 tbsp Cinamon",
			"2 tbsp Chia seeds",
			"2 tbsp Flax seed",
			" Peanut butter"
		],
		"recipeInstructions": [
			{
				"@type": "HowToStep",
				"text": "Put everything in the blinder except peanut butter."
			},
			{
				"@type": "HowToStep",
				"text": "Blend like it's nobody's business."
			},
			{ "@type": "HowToStep", "text": "Put in container." },
			{ "@type": "HowToStep", "text": "Drizzle your peanut butter on it." },
			{
				"@type": "HowToStep",
				"text": "Make some design lines with a nife (it's to mix the PB lightly)."
			},
			{ "@type": "HowToStep", "text": "Leave it until tomorrow." }
		],
		"image": "https://storage.googleapis.com/static.4ks.io/fallback/f0.jpg"
	}`

	var p LdProcessor
	var out interface{}

	// unmarshal json
	var data map[string]interface{}
	err := json.Unmarshal([]byte(recipe), &data)
	if err != nil {
		panic(err)
	}

	// notice no network calls are performed
	p = NewLdProcessor(client, true)
	out, err = p.Parse(data)
	if err != nil {
		// handle error
	}
	PrintStruct(out)

	// notice the network call to retrieve the schema
	disableCaching := false
	p = NewLdProcessor(client, disableCaching)
	out, err = p.Parse(data)
	if err != nil {
		// handle error
	}
	PrintStruct(out)
}

// PrintStruct prints a struct
func PrintStruct(t interface{}) {
	j, _ := json.MarshalIndent(t, "", "  ")
	fmt.Println(string(j))
}
