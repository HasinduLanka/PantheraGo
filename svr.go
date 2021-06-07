package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type API_POST_Recieved func([]string, []byte) Response
type API_GET_Recieved func([]string) Response

type Response struct {
	body    []byte
	headers map[string]string
}

func BodyResponse(body []byte) Response {
	return Response{body, map[string]string{}}
}
func StringResponse(body string) Response {
	return Response{[]byte(body), map[string]string{}}
}

var API_POSTs map[*regexp.Regexp]API_POST_Recieved = map[*regexp.Regexp]API_POST_Recieved{
	regexp.MustCompile(`^api/(.*)`): APIPostCall}

var API_GETs map[*regexp.Regexp]API_GET_Recieved = map[*regexp.Regexp]API_GET_Recieved{
	regexp.MustCompile(`^.*\.go\.html`):                 ServePanthera,
	regexp.MustCompile(`^api/goevent/(.*)/(.*)/(.*)/*`): APIGoEvent}

func HttpServe() {
	FullMux := http.NewServeMux()
	FullMux.HandleFunc("/", ServeAPIsAndFiles)

	println("Serving on http://localhost:56749 ")

	if err := http.ListenAndServe(":56749", FullMux); err != nil {
		log.Fatal(err)
	}
}

func ServeAPIsAndFiles(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/" {
	// 	http.Error(w, "404 not found.", http.StatusNotFound)
	// 	return
	// }

	urlpath := strings.ReplaceAll(strings.TrimPrefix(r.URL.Path, "/"), "..", "")

	switch r.Method {
	case "GET":

		if len(urlpath) == 0 || urlpath == "index" || urlpath == "index.html" {
			urlpath = "index.go.html"
		}

		for apiPath, api := range API_GETs {
			Matches := apiPath.FindAllStringSubmatch(urlpath, 8)
			if len(Matches) != 0 {
				fmt.Println("Serving API " + strings.Join(Matches[0], " - "))
				resp := api(Matches[0])
				ReturnResponse(resp, w)
				return
			}

		}

		// If no API
		fmt.Println("Serving file " + urlpath)
		http.ServeFile(w, r, urlpath)

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		println("POST " + urlpath + " body length " + strconv.Itoa(len(body)))

		for apiPath, api := range API_POSTs {
			Matches := apiPath.FindAllStringSubmatch(urlpath, 8)
			if len(Matches) != 0 {
				fmt.Println("Serving API " + strings.Join(Matches[0], " - "))
				resp := api(Matches[0], body)
				ReturnResponse(resp, w)
				return
			}

		}

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func ReturnResponse(resp Response, w http.ResponseWriter) {
	if resp.headers != nil {
		for k, v := range resp.headers {
			w.Header().Set(k, v)
		}
	}
	if len(resp.body) > 0 {
		w.Write(resp.body)
	}
}

func APIPostCall(urlPath []string, body []byte) Response {
	return Response{}
}

// api/id/sender/para -> ReloadRequired?
func APIGoEvent(urlPath []string) Response {
	if len(urlPath) != 4 {
		println("API path error : " + strings.Join(urlPath, " - "))
		return StringResponse("error:500")
	}

	evrsp := DefaultSession.CallEvent(urlPath[1], urlPath[2], urlPath[3])
	brsp, jerr := json.Marshal(evrsp)
	if jerr != nil {
		PrintError(jerr)
	}

	return BodyResponse(brsp)
}
