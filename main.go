package main

import (
	"fmt"
	"strconv"
	"time"
)

var DefaultSession Session
var ComponentSrcs map[string]*ComponentSrc

func main() {
	fmt.Println("Panthera.Go")

	ComponentSrcs = map[string]*ComponentSrc{
		"dash":    ComponentSrc{}.Make("dash", URIProvider("dash.go.html")),
		"footer":  ComponentSrc{}.Make("footer", URIProvider("footer.go.html")),
		"counter": ComponentSrc{}.Make("counter", URIProvider("counter.go.html")).SetVarsOnNew(map[string]string{"count": "100"}),
	}

	DefaultSession = Session{
		ID:                   "Default",
		ComponentSrcProvider: ComponentSrcProvider,
	}

	DefaultSession.MakeRoot()

	DefaultSession.SetVar("root.var1", "This variable is available throughout the session")
	DefaultSession.SetVar("root.company-name", "Bitblazers")
	// DefaultSession.RootComp.Src.SetFunc()
	DefaultSession.RootComp.Src.SetFunc("time", func(c *Component) string {
		return time.Now().String()
	})

	DefaultSession.SetVar("dash1.company-name", "Bitblazers")

	DefaultSession.SetVar("dash1.bigtext", "PantheraGo is a HTML DOM manipulator written in <strong> Go Language. </strong> PantheraGo can run in the browser using Web Assembly and also, it can function as a webserver running on Linux, Mac, Windows and Android natively ")

	DefaultSession.SetVar("dash1.myvar", "123")
	DefaultSession.SetVar("dash1.count", "145")

	Dash := ComponentSrcProvider("dash")

	Dash.SetFunc("new", func(c *Component) string {
		c.SetVar("dash1.count", "123")
		return ""
	})

	DefaultSession.ResolveCompPath("dash1").Src = Dash

	// DefaultSession.SetFunc("dash.sayhello", SayTime)

	// DefaultSession.ResolveCompPath("dash").r

	// DefaultSession.SetEvent("root.btn-count-click", BtnCountClick)
	// DefaultSession.SetEvent("root.txt-name-change", TxtNameChanged)

	HttpServe()
}

func URIProvider(uri string) func() string {
	return func() string {
		HTML, _ := LoadURIToString(uri)
		return HTML
	}
}

func ComponentSrcProvider(src string) *ComponentSrc {
	C, found := ComponentSrcs[src]
	if !found {
		return nil
	}
	return C
}

func ServePanthera(Args []string) Response {
	HTML, _ := LoadURIToString(Args[0])
	Panthtml := DefaultSession.Render(HTML, DefaultSession.RootComp)
	return StringResponse(Panthtml)
}

func SayTime() string {

	return "Hello there. Time is " + time.Now().String()
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
