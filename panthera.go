package main

import (
	"regexp"
	"strings"
)

var RegexGo *regexp.Regexp = regexp.MustCompile("(?s)< *go  *(.*?)>(.*?)< */go *>")
var RegexVar *regexp.Regexp = regexp.MustCompile(`var *= *"(.*?)".*?>`)
var RegexFunc *regexp.Regexp = regexp.MustCompile(`func *= *"(.*?)"(.*?)>`)
var RegexArgs *regexp.Regexp = regexp.MustCompile(` +(.*) *= *"(.*?)"`)

var NewLineRemover *strings.Replacer = strings.NewReplacer("\n", "")

var GoVars map[string]string = map[string]string{}
var GoFuncs map[string]func(map[string]string) string = map[string]func(map[string]string) string{}
var GoEvents map[string]func(string, string) EventResponse = map[string]func(string, string) EventResponse{}

type EventResponse struct {
	Reload  bool
	Update  bool
	ID      string
	Content string
}

func Render(B string) string {

	return RegexGo.ReplaceAllStringFunc(B, renderTag)

}

func RenderReusable(B string) string {

	return RegexGo.ReplaceAllStringFunc(B, renderReusableTag)

}

func renderTag(T string) string {
	MatchesVar := RegexVar.FindStringSubmatch(T)
	if len(MatchesVar) > 1 {
		return renderVar(MatchesVar[1])
	}

	MatchesFunc := RegexFunc.FindStringSubmatch(T)
	if len(MatchesFunc) == 2 {
		return renderFunc(MatchesFunc[1], "")
	} else if len(MatchesFunc) > 2 {
		return renderFunc(MatchesFunc[1], MatchesFunc[2])
	}

	return T
}

func renderReusableTag(T string) string {
	MatchesVar := RegexVar.FindStringSubmatch(T)
	if len(MatchesVar) > 1 {
		return `<go ` + MatchesVar[0] + renderVar(MatchesVar[1]) + ` </go> `
	}

	MatchesFunc := RegexFunc.FindStringSubmatch(T)
	if len(MatchesFunc) == 2 {
		return `<go ` + MatchesFunc[0] + renderFunc(MatchesFunc[1], "") + `</go> `
	} else if len(MatchesFunc) > 2 {
		return `<go ` + MatchesFunc[0] + renderFunc(MatchesFunc[1], MatchesFunc[2]) + `</go> `
	}

	return T
}

func renderVar(T string) string {
	return `<a id="` + T + `">` + GetVar(T) + `</a> `
}

func renderFunc(T string, Args string) string {
	return CallFunc(T, Args)
}

func SetVar(name string, val string) {
	GoVars[name] = val
}

func GetVar(name string) string {
	val, found := GoVars[name]
	if !found {
		return ""
	}
	return val
}

func SetFunc(name string, val func(map[string]string) string) {
	GoFuncs[name] = val
}

func CallFunc(name string, Args string) string {
	fn, found := GoFuncs[name]
	if !found {
		return ""
	}

	MArgs := RegexArgs.FindAllStringSubmatch(Args, 8)
	FArgs := make(map[string]string, len(MArgs))
	if len(MArgs) > 0 {
		for _, arg := range MArgs {
			if len(arg) == 3 {
				FArgs[arg[1]] = arg[2]
			}
		}
	}

	return fn(FArgs)

}

// func(sender, para) ReloadRequired?
func SetEvent(id string, val func(string, string) EventResponse) {
	GoEvents[id] = val
}

func CallEvent(id string, sender string, para string) EventResponse {
	ev, found := GoEvents[id]
	if !found {
		return EventResponse{Reload: false, Update: false}
	}

	if len(sender) == 0 {
		sender = "nosender"
	}

	if len(para) == 0 {
		para = "null"
	}

	return ev(sender, para)
}
