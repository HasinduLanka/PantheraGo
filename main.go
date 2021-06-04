package main

import (
	"bytes"
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("Panthera.Go")

	SetVar("company-name", "Bitblazers")

	SetVar("bigtext", "PantheraGo is a HTML DOM manipulator written in <strong> Go Language. </strong> PantheraGo can run in the browser using Web Assembly and also, it can function as a webserver running on Linux, Mac, Windows and Android natively ")

	SetVar("myvar", "123")
	SetVar("count", "1")

	SetFunc("sayhello", SayHello)

	SetEvent("btn-count-click", BtnCountClick)
	SetEvent("txt-name-change", TxtNameChanged)

	HttpServe()
}

func ServePanthera(Args []string) Response {
	HTML, _ := LoadURIToString(Args[0])
	Panthtml := Render(HTML)
	return StringResponse(Panthtml)
}

func SayHello(args map[string]string) string {
	name, found := args["name"]
	if found {
		return "Hello there " + name
	}
	return "Please tell me your name"
}

func BtnCountClick(sender, Para string) EventResponse {
	println("Event raised : From " + sender + " :  Para " + Para)
	count, err := strconv.Atoi(GetVar("count"))
	if err != nil {
		count = 0
	}
	count++

	SetVar("count", strconv.Itoa(count))
	SetVar("myvar", "Whole document can be reloaded on each event "+strconv.Itoa(count))

	return EventResponse{Reload: true}
}

func TxtNameChanged(sender, Para string) EventResponse {
	if len(Para) == 0 {
		return EventResponse{Reload: false, Update: false}
	}

	var ch chan string = make(chan string)
	go BuildSomeLongText(Para, ch)

	var buffer bytes.Buffer

	for perm := range ch {
		buffer.WriteString(perm)
	}

	R := buffer.String()
	SetVar("name-result", R)
	return EventResponse{Reload: false, Update: true, ID: "name-result", Content: R}
}

func BuildSomeLongText(s string, ch chan string) {
	ch <- " <br>"
	for i := 0; i < len(s)*10; i++ {
		ch <- s + " " + strconv.Itoa(i) + " <br>"
	}
	close(ch)
}
