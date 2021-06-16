// Copyright (c) 2021, Microsoft Corporation, Sean Hinchee
// Licensed under the MIT License.

package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Response represents an HTTP response - omit the interface from "http"
type Response struct {
	Status           string
	StatusCode       int
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	Body             string
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Uncompressed     bool
	TLS              *tls.ConnectionState
}

var (
	chatty  = flag.Bool("D", false, "verbose debug output")
	toJSON  = flag.Bool("j", false, "emit response as JSON")
	yesTLS  = flag.Bool("T", false, "populate response TLS information (JSON)")
	noBody  = flag.Bool("B", false, "omit the body in the response (JSON)")
	inFile  = flag.String("i", "", "file to read request from (if not stdin)")
	subBody = flag.String("b", "", "substitute request body, if any")
	b64     = flag.Bool("d", false, "is the substitute body base64-encoded?")
	proto   = flag.String("p", "https", "protocol to use for request")
)

// Send a specifically crafted HTTP by reading from stdin or a file
func main() {
	flag.Parse()
	args := flag.Args()

	// Scan for substitutions
	// Headers as `foo: bar`
	// Query tuples as `foo? bar`
	queries := make(map[string]string)
	headers := make(map[string]string)

	for _, arg := range args {
		// Might be unsafe ☺
		key := strings.Fields(arg)[0]
		if strings.ContainsRune(key, '?') {
			// Query
			key = key[:len(key)-1]

			queries[key] = strings.SplitN(arg, "? ", 2)[1]
			continue
		}

		if strings.ContainsRune(key, ':') {
			// Header
			key = key[:len(key)-1]

			headers[key] = strings.SplitN(arg, ": ", 2)[1]
			continue
		}

		// Invalid format
		fatal("err: header/query substitutions must have the first word end in '?' or ':'")
	}

	// Set input/output
	var in io.Reader = os.Stdin
	out := bufio.NewWriter(os.Stdout)

	if *inFile != "" {
		f, err := os.Open(*inFile)
		if err != nil {
			fatal("err: could not open file", *inFile, "-", err)
		}
		in = f
	}

	req, err := http.ReadRequest(bufio.NewReader(in))
	if err != nil {
		fatal("err: could not parse http request -", err)
	}

	// Patching to make requests pass
	req.RequestURI = ""
	req.URL.Scheme = *proto
	req.URL.Host = req.Host

	// Override HTTP headers
	for k, v := range headers {
		req.Header[k] = []string{v}
	}

	// Override URL query tuples
	vals := req.URL.Query()
	for k, v := range queries {
		vals[k] = []string{v}
	}
	req.URL.RawQuery = vals.Encode()

	// Override body
	if *subBody != "" {
		if *b64 {
			buf, err := base64.StdEncoding.DecodeString(*subBody)
			if err != nil {
				fatal("err: could not base64-decode body")
			}

			*subBody = string(buf)
		}
		req.Body = ioutil.NopCloser(strings.NewReader(*subBody))
	}

	client := &http.Client{}

	if *chatty {
		fmt.Printf("%#v\n\n", req)
		fmt.Printf("%#v\n\n", req.URL)
	}

	resp, err := client.Do(req)
	if err != nil {
		fatal("err: could not make request -", err)
	}

	if *toJSON {
		r := http2response(*resp)
		enc := json.NewEncoder(out)
		err := enc.Encode(r)
		if err != nil {
			fatal("err: could not encode response to JSON -", err)
		}
		out.Flush()
		return
	}

	resp.Write(out)
	out.WriteRune('\n') // Maybe make this optional?
	out.Flush()
}

// Fatal - end program with an error message and newline
func fatal(s ...interface{}) {
	fmt.Fprintln(os.Stderr, s...)
	os.Exit(1)
}

func http2response(r http.Response) Response {
	resp := Response{
		Status:           r.Status,
		StatusCode:       r.StatusCode,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header,
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close:            r.Close,
		Uncompressed:     r.Uncompressed,
	}

	if !*noBody {
		var buf bytes.Buffer
		buf.ReadFrom(r.Body)
		body := buf.String()
		resp.Body = body
	}

	if *yesTLS {
		resp.TLS = r.TLS
	}

	return resp
}
