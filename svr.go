package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type API_POST_Recieved func(*SvrRequest, []byte) Response
type API_GET_Recieved func(*SvrRequest) Response

type Response struct {
	body    []byte
	headers map[string]string
}

type SvrRequest struct {
	Paths         []string
	Req           *http.Request
	Resp          *http.ResponseWriter
	Body          []byte // nil for GET
	session_cache *Session
}

func BodyResponse(body []byte) Response {
	return Response{body, map[string]string{}}
}
func StringResponse(body string) Response {
	return Response{[]byte(body), map[string]string{}}
}

func (sv *SvrRequest) AuthSession() *Session {

	if sv.session_cache != nil {
		return sv.session_cache
	}

	if sv.Req == nil {
		return nil
	}

	sidck, ckerr := sv.Req.Cookie("panthera-session")
	passck, passerr := sv.Req.Cookie("panthera-pass")

	var S *Session = nil

	if ckerr == nil && passerr == nil {
		sid := sidck.Value
		pass := passck.Value
		S = SessionMngr.Get(sid, pass)

	}

	if S == nil {
		S = SessionMngr.NewSession()
		(*sv.Resp).Header().Add("panthera-new", "true")

	}

	sv.SetSessionCookies(S)

	sv.session_cache = S
	return S
}

func (sv *SvrRequest) SetSessionCookies(S *Session) {
	http.SetCookie(*sv.Resp, &http.Cookie{
		Name:   "panthera-session",
		Value:  S.ID,
		MaxAge: 60 * 30,
	})
	http.SetCookie(*sv.Resp, &http.Cookie{
		Name:   "panthera-pass",
		Value:  S.Pass,
		MaxAge: 60 * 30,
	})
}

func (sv *SvrRequest) NewSession() {

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
				resp := api(&SvrRequest{Paths: Matches[0], Req: r, Resp: &w})
				svrReturnResponse(resp, w)
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
				resp := api(&SvrRequest{Paths: Matches[0], Req: r, Resp: &w}, body)
				svrReturnResponse(resp, w)
				return
			}

		}

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func svrReturnResponse(resp Response, w http.ResponseWriter) {
	if resp.headers != nil {
		for k, v := range resp.headers {
			w.Header().Set(k, v)
		}
	}
	if len(resp.body) > 0 {
		w.Write(resp.body)
	}
}

func APIPostCall(Req *SvrRequest, body []byte) Response {
	return Response{}
}
