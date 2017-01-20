package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"bytes"
	"flag"
	"encoding/base64"
)

type UserRequest struct {
	Url     string                   `json:"url"`
	Method  string                   `json:"method"`
	Headers map[string]string        `json:"headers"`
	Body    string                   `json:"body"`
}

type Reader struct {
	io.Reader
	io.Closer
}

var (
	debug bool
	auth  string
)

func (r *Reader) Close() error {
	return nil
}

func passthrough(request *http.Request, writer http.ResponseWriter, content io.ReadCloser) error {
	userReq := UserRequest{}
	decoder := json.NewDecoder(content)
	err := decoder.Decode(&userReq)
	if err != nil {
		return err
	}

	//send http request
	httpReq, err := generateRequest(&userReq)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	} else if response.Body != nil {
		defer response.Body.Close()
	}

	//copy headers
	headers := writer.Header()
	for k, v := range response.Header {
		headers.Add(k, v[0])
	}

	//do the dance
	corsDance(request, writer)
	writer.WriteHeader(response.StatusCode)

	//write body
	var buff [512]byte
	io.CopyBuffer(writer, response.Body, buff[:])

	//hooray!
	return nil
}

func generateRequest(userReq *UserRequest) (*http.Request, error) {
	//convert body to reader
	reqBody := &Reader{Reader:bytes.NewReader([]byte(userReq.Body))}

	//convert header map into string array
	headers := make(map[string][]string)
	for k, v := range userReq.Headers {
		headers[k] = []string{v}
	}

	//add auth header
	if len(auth) > 0 {
		headers["Authorization"] = []string{auth}
	}

	//generate request
	request, err := http.NewRequest(userReq.Method, userReq.Url, reqBody)
	if err != nil {
		return nil, err
	}
	request.Header = headers

	if debug {
		str := fmt.Sprintf("%v: %v\n", userReq.Method, userReq.Url)
		for k, v := range headers {
			str += fmt.Sprintf("\t%v: %v\n", k, v[0])
		}
		fmt.Print(str)
	}
	return request, nil
}

func corsDance(request *http.Request, writer http.ResponseWriter) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	writer.Header().Set("Access-Control-Request-Method", request.Header.Get("Access-Control-Request-Method"))
}

func main() {
	debug_flag := flag.Bool("debug", false, "Enable debug output")
	address := flag.String("address", "0.0.0.0", "The address the server will bind to")
	port := flag.String("port", "8080", "The port the server will bind to")
	auth_flag := flag.String("auth", "", "Enable Basic Auth on every request in user:pass form")

	flag.Parse()
	debug = *debug_flag
	if *auth_flag != "" {
		auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(*auth_flag))
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		var err error
		if request.Method == "POST" {
			err = passthrough(request, writer, request.Body)
		} else if request.Method == "OPTIONS" {
			corsDance(request, writer)
			writer.WriteHeader(200)
		} else {
			writer.WriteHeader(400)
			writer.Write([]byte(`Invalid request method`))
			return
		}

		if err != nil {
			corsDance(request, writer)
			writer.WriteHeader(500)
			writer.Write([]byte(err.Error()))
			fmt.Printf("Error! '%v'\n", err)
		}
	})

	err := http.ListenAndServe((*address)+":"+(*port), nil)
	fmt.Printf("%v\n", err)
}
