package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var SessionMngr SessionManager

func main() {
	fmt.Println("Panthera.Go")

	MakeDefaultSession()

	HttpServe()
}

func MakeDefaultSession() {

	SessionMngr = SessionManager{Sessions: map[string]*Session{}}
	SessionMngr.Make()

	SessionMngr.ComponentSrcs = map[string]*ComponentSrc{
		// Root is a virtual component at Session.RootComp of every Session. Think it as a global variable store for Session
		"root":    ComponentSrc{}.Make("root", func() string { return "" }), // No need of source for Root
		"dash":    ComponentSrc{}.Make("dash", URIProvider("dash.go.html")),
		"footer":  ComponentSrc{}.Make("footer", URIProvider("footer.go.html")),
		"counter": ComponentSrc{}.Make("counter", URIProvider("counter.go.html")).SetVarsOnNew(map[string]string{"count": "100"}),
	}

	DefaultSession := Session{ID: "Default", ComponentSrcProvider: SessionMngr.ComponentSrcProvider}
	DefaultSession.MakeRoot()

	DefaultSession.SetVar("root.var1", "This variable is available throughout the session")
	DefaultSession.SetVar("root.company-name", "Bitblazers")

	DefaultSession.SetVar("root.dash1.company-name", "Bitblazers")

	DefaultSession.SetVar("root.dash1.bigtext", "PantheraGo is a HTML DOM manipulator written in <strong> Go Language. </strong> PantheraGo can run in the browser using Web Assembly and also, it can function as a webserver running on Linux, Mac, Windows and Android natively ")

	DefaultSession.SetVar("root.dash1.myvar", "123")

	RootSrc := SessionMngr.ComponentSrcProvider("root")
	RootSrc.SetFunc("time", func_Root_Time)

	// DefaultSession.SetFunc("dash.sayhello", SayTime)

	// DefaultSession.ResolveCompPath("dash").r

	// DefaultSession.SetEvent("root.btn-count-click", BtnCountClick)
	// DefaultSession.SetEvent("root.txt-name-change", TxtNameChanged)

	SessionMngr.DefaultSessionStash = DefaultSession.Stash()

}

func URIProvider(uri string) func() string {
	return func() string {
		HTML, _ := LoadURIToString(uri)
		return HTML
	}
}

func ServePanthera(Req *SvrRequest) Response {
	HTML, _ := LoadURIToString(Req.Paths[0])

	CurrentSession := Req.AuthSession()
	CurrentSession.SetVar("root.panthera-session", CurrentSession.ID)

	Panthtml := CurrentSession.RootComp.RenderSource(HTML)
	return StringResponse(Panthtml)
}

func APIGoEvent(Req *SvrRequest) Response {
	if len(Req.Paths) != 4 {
		println("API path error : " + strings.Join(Req.Paths, " - "))
		return StringResponse("error:500")
	}

	CurrentSession := Req.AuthSession()

	if CurrentSession == nil {
		println("Error. No Session")
		return StringResponse("NoSession")
	}

	evrsp := CurrentSession.CallEvent(Req.Paths[1], Req.Paths[2], Req.Paths[3])
	brsp, jerr := json.Marshal(evrsp)
	if jerr != nil {
		PrintError(jerr)
	}

	return BodyResponse(brsp)
}

func func_Root_Time(c *Component) string {
	return time.Now().String()
}

// func BtnCountClick(sender, Para string) EventResponse {
// 	println("Event raised : From " + sender + " :  Para " + Para)
// 	count, err := strconv.Atoi(GetVar("count"))
// 	if err != nil {
// 		count = 0
// 	}
// 	count++

// 	SetVar("count", strconv.Itoa(count))
// 	SetVar("myvar", "Whole document can be reloaded on each event "+strconv.Itoa(count))

// 	return EventResponse{Reload: true}
// }

// func TxtNameChanged(sender, Para string) EventResponse {
// 	if len(Para) == 0 {
// 		return EventResponse{Reload: false, Update: false}
// 	}

// 	var ch chan string = make(chan string)
// 	go BuildSomeLongText(Para, ch)

// 	var buffer bytes.Buffer

// 	for perm := range ch {
// 		buffer.WriteString(perm)
// 	}

// 	R := buffer.String()
// 	SetVar("name-result", R)
// 	return EventResponse{Reload: false, Update: true, ID: "name-result", Content: R}
// }

func BuildSomeLongText(s string, ch chan string) {
	ch <- " <br>"
	for i := 0; i < len(s)*10; i++ {
		ch <- s + " " + strconv.Itoa(i) + " <br>"
	}
	close(ch)
}
