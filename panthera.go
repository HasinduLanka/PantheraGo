package main

import (
	"regexp"
	"sort"
	"strings"
)

var RegexGoCompBegin *regexp.Regexp = regexp.MustCompile(`< *gocomp +(.*?)>`)
var RegexGoCompEnd *regexp.Regexp = regexp.MustCompile(`< *?/gocomp *?>`)
var RegexGo *regexp.Regexp = regexp.MustCompile(`(?s)< *go +(.*?)>(.*?)< */go *>`)
var RegexVar *regexp.Regexp = regexp.MustCompile(` var *= *"(.*?)".*?>`)
var RegexFunc *regexp.Regexp = regexp.MustCompile(` func *= *"(.*?)"(.*?)>`)
var RegexArgs *regexp.Regexp = regexp.MustCompile(` *(.*?) *= *"(.*?)" *`)
var RegexGoAttrs *regexp.Regexp = regexp.MustCompile(`<.*? goattributes *= *"true" .*?>`)
var RegexGoAttr *regexp.Regexp = regexp.MustCompile(` goattr-(.*?) *= *"(.*?)"`)

var NewLineRemover *strings.Replacer = strings.NewReplacer("\n", "")
var replacerEscapeForRegex *strings.Replacer = strings.NewReplacer(`.`, `\.`, `*`, `\*`, `\`, `\\`)

func EscapeForRegex(s string) string {
	return replacerEscapeForRegex.Replace(s)
}

// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //

type Session struct {
	ID                   string
	RootComp             *Component
	ComponentSrcProvider func(string) *ComponentSrc
}

func (S *Session) Render(HTML string, Parent *Component) string {

	CB := RegexGoCompBegin.FindAllStringSubmatchIndex(HTML, 4096)
	L := len(CB)
	if L == 0 {
		Parent.ChildComps = map[string]*Component{}
		return HTML
	}

	CE := RegexGoCompEnd.FindAllStringSubmatchIndex(HTML, 4096)
	if L != len(CE) {
		return "Component Begin/End Pairs malformed"
	}

	type CompDetection struct {
		IsBegin bool
		Arr     []int
		Args    map[string]string
	}

	CompIndexes := make(map[int]CompDetection, L*2)
	CompIndexArr := make([]int, L*2)
	iCompIndexArr := 0

	for _, v := range CB {
		sargs := HTML[v[2]:v[3]]
		AArgs := RegexArgs.FindAllStringSubmatch(sargs, 256)
		Args := make(map[string]string, len(AArgs))

		for _, aarg := range AArgs {
			Args[strings.TrimPrefix(aarg[1], "gopara-")] = aarg[2]
		}

		CompIndexes[v[0]] = CompDetection{IsBegin: true, Arr: v, Args: Args}
		CompIndexArr[iCompIndexArr] = v[0]
		iCompIndexArr++
	}
	for _, v := range CE {
		CompIndexes[v[0]] = CompDetection{IsBegin: false, Arr: v}
		CompIndexArr[iCompIndexArr] = v[0]
		iCompIndexArr++
	}

	sort.Ints(CompIndexArr)
	depth := 0

	Comps := map[string]*Component{}
	CompsRanges := map[string][]int{} // Begin2, End2
	CompList := []string{}

	LHTML := len(HTML)

	CompID := "#"
	for i := 0; i < L*2; i++ {
		ci := CompIndexArr[i]
		CI := CompIndexes[ci]

		if CI.IsBegin {

			if depth == 0 {
				// CompID := HTML[CI.Arr[2]:CI.Arr[3]]
				CompID = CI.Args["compid"]
				CompSrcs := CI.Args["compsrc"]

			CheckDuplicateIDs:
				_, found := Comps[CompID]
				if found {
					CompID += "-copy"
					goto CheckDuplicateIDs
				}

				var C *Component

				OldC, Old := Parent.ChildComps[CompID]
				if Old {
					if OldC.Src == nil || OldC.Src.SrcID != CompSrcs {
						C = OldC
					} else {
						OldC.SetVars(CI.Args)
						Comps[CompID] = OldC
						CompsRanges[CompID] = []int{CI.Arr[1], LHTML}
						CompList = append(CompList, CompID)
						depth++
						continue
					}
				} else {
					C = &Component{
						Path:       Parent.Path + "." + Parent.ID,
						ID:         CompID,
						Session:    S,
						Parent:     Parent,
						ChildComps: map[string]*Component{},
						GoVars:     map[string]string{}}
				}

				CompSrc := S.ComponentSrcProvider(CompSrcs)
				if CompSrc == nil {
					println("Component source not found. ID:" + CompID + " Src:" + CompSrcs)
				}

				C.Src = CompSrc
				C.CallEvent("new", "", "")
				C.SetVars(CI.Args)

				Comps[CompID] = C
				CompsRanges[CompID] = []int{CI.Arr[1], LHTML}
				CompList = append(CompList, CompID)

			}
			depth++
		} else {
			depth--

			if depth == 0 {
				if CompID != "#" && len(CompsRanges[CompID]) == 2 {
					CompsRanges[CompID][1] = CI.Arr[1]
				}
			}
		}

	}

	Parent.ChildComps = Comps

	var sb strings.Builder
	LastI := 0
	for _, v := range CompList {
		CompRng := CompsRanges[v]
		if LastI < CompRng[0] { // Double IDs
			sb.WriteString(HTML[LastI:CompRng[0]] + "\n")
			C := Comps[v]
			sb.WriteString(C.Render())
			sb.WriteString("\n</gocomp>\n")
			LastI = CompRng[1]
		}
	}

	sb.WriteString(HTML[LastI:])

	return sb.String()

}

func (S *Session) ResolveCompPath(path string) *Component {
	p := strings.Split(path, ".")
	return S.ResolveComp(p)
}

func (S *Session) ResolveComp(p []string) *Component {

	if len(p) == 0 {
		return S.RootComp
	}

	var C *Component = S.RootComp
	var starti int

	if len(p[0]) == 0 || p[0] == S.RootComp.ID {
		starti = 1
	} else {
		starti = 0
	}

	for i := starti; i < len(p); i++ {
		ch, found := C.ChildComps[p[i]]
		if found {
			C = ch
		} else {
			ch = &Component{
				Path:       strings.Join(p[:i], "."),
				ID:         p[i],
				Src:        nil,
				Session:    S,
				Parent:     C,
				ChildComps: map[string]*Component{},
				GoVars:     map[string]string{},
			}

			C.ChildComps[p[i]] = ch
			C = ch
		}
	}

	return C
}

func (S *Session) MakeRoot() {
	S.RootComp = &Component{
		Path: "",
		ID:   "root",
		Src: &ComponentSrc{
			SrcID: "root",
			Provider: func() string {
				return ""
			},
			GoFuncs:  map[string]func(*Component) string{},
			GoEvents: map[string]func(*Component, string, string) EventResponse{},
		},
		Session:    &DefaultSession,
		Parent:     nil,
		ChildComps: map[string]*Component{},
		GoVars:     map[string]string{},
	}
}

// "root.comp1.var1" -> (comp1, "var1")
func (S *Session) ComponentOf(path string) (*Component, string) {
	P := strings.Split(path, ".")
	lenp := len(P)
	if lenp == 0 {
		return nil, ""
	}

	C := S.ResolveComp(P[:lenp-1])

	if lenp == 1 {
		return C, ""
	}

	return C, P[lenp-1]

}

func (S *Session) SetVar(path string, val string) {

	C, name := S.ComponentOf(path)
	if C == nil {
		return
	}

	C.SetVar(name, val)
}

func (S *Session) GetVar(path string) string {
	C, name := S.ComponentOf(path)
	if C == nil {
		return ""
	}

	return C.GetVar(name)
}

func (S *Session) CallFunc(path string) string {
	C, name := S.ComponentOf(path)
	if C == nil {
		return ""
	}

	return C.CallFunc(name)
}

func (S *Session) CallEvent(path string, sender string, para string) EventResponse {
	C, name := S.ComponentOf(path)
	if C == nil {
		return EventResponse{}
	}

	return C.CallEvent(name, sender, para)
}

// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //

// AKA ComponentType
type ComponentSrc struct {
	SrcID    string
	Provider func() string `json:"-"`

	GoFuncs  map[string]func(*Component) string                        `json:"-"`
	GoEvents map[string]func(*Component, string, string) EventResponse `json:"-"`
}

func (T ComponentSrc) Make(SrcID string, Provider func() string) *ComponentSrc {
	T.SrcID = SrcID
	T.Provider = Provider
	T.GoFuncs = map[string]func(*Component) string{}
	T.GoEvents = map[string]func(*Component, string, string) EventResponse{}

	return &T
}

func (T *ComponentSrc) SetVarsOnNew(vars map[string]string) *ComponentSrc {
	New := func(C *Component, sender string, para string) EventResponse {
		C.SetVars(vars)
		return EventResponse{}
	}

	OldNew, OldNewFound := T.GoEvents["new"]
	if OldNewFound {
		T.GoEvents["new"] = func(c *Component, s1, s2 string) EventResponse {
			New(c, s1, s2)
			return OldNew(c, s1, s2)
		}
	} else {
		T.GoEvents["new"] = New
	}

	return T
}

// func (T *ComponentSrc) SetEvent(name string, ev func(*Component, string, string) EventResponse) {
// 	T.GoEvents[name] = ev
// }

// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //

type Component struct {
	Path    string
	ID      string
	Src     *ComponentSrc
	Session *Session `json:"-"`

	Parent     *Component            `json:"-"`
	ChildComps map[string]*Component `json:"-"`

	GoVars map[string]string
}

type EventResponse struct {
	Update DOMEffect
}

func (C *Component) Render() string {
	if C.Src == nil {
		return "No source for component " + C.ID
	}
	S2 := C.Session.Render(C.Src.Provider(), C)

	S2 = RegexGoAttrs.ReplaceAllStringFunc(S2,
		func(s string) string {
			matches := RegexGoAttr.FindAllStringSubmatch(s, 64)
			s = strings.TrimSuffix(s, ">")

			for _, match := range matches {
				if len(match) == 3 {
					kattr := EscapeForRegex(match[1])
					vattr := EscapeForRegex(match[2])
					re, rerr := regexp.Compile(` ` + kattr + ` *= *".*?"`)
					if rerr != nil {
						println("Go Attribute syntax error : " + kattr + `="` + vattr + `"`)
					}
					s = re.ReplaceAllString(s, "")

					s += ` ` + kattr + `="` + C.CallFunc(match[2]) + `"`
				}
			}
			s += " >"
			return s
		})

	return RegexGo.ReplaceAllStringFunc(S2, C.renderTag)
}

func (C *Component) RenderReusable() string {
	if C.Src == nil {
		return "No source for reusable component " + C.ID
	}
	S2 := C.Session.Render(C.Src.Provider(), C)
	return RegexGo.ReplaceAllStringFunc(S2, C.renderReusableTag)
}

func (C *Component) renderTag(T string) string {
	MatchesVar := RegexVar.FindStringSubmatch(T)
	if len(MatchesVar) > 1 {
		return C.renderVar(MatchesVar[1])
	}

	MatchesFunc := RegexFunc.FindStringSubmatch(T)
	if len(MatchesFunc) > 1 {
		return C.renderFunc(MatchesFunc[1])
	}

	return T
}

func (C *Component) renderReusableTag(T string) string {
	MatchesVar := RegexVar.FindStringSubmatch(T)
	if len(MatchesVar) > 1 {
		return `<go ` + MatchesVar[0] + C.renderVar(MatchesVar[1]) + ` </go> `
	}

	MatchesFunc := RegexFunc.FindStringSubmatch(T)
	if len(MatchesFunc) > 1 {
		return `<go ` + MatchesFunc[0] + C.renderFunc(MatchesFunc[1]) + `</go> `
	}

	return T
}

func (C *Component) renderVar(T string) string {
	Pathed := strings.Contains(T, ".")
	var R string
	if Pathed {
		R = C.Session.GetVar(T)
	} else {
		R = C.GetVar(T)
	}
	return `<span govar="` + T + `">` + R + `</span> `
}

func (C *Component) renderFunc(T string) string {
	return `<span gofunc="` + T + `">` + C.CallFunc(T) + `</span> `
}

func (C *Component) SetVar(name string, val string) {

	Pathed := strings.Contains(name, ".")

	if Pathed {
		if C.ChildComps != nil && len(C.ChildComps) != 0 {
			splts := strings.SplitN(name, ".", 2)
			ch, chf := C.ChildComps[splts[0]]
			if chf {
				ch.SetVar(splts[1], val)
				return
			}
		}
		C.Session.SetVar(name, val)
	} else {
		C.GoVars[name] = val
	}

}

func (C *Component) GetVar(name string) string {
	Pathed := strings.Contains(name, ".")

	if Pathed {
		if C.ChildComps != nil && len(C.ChildComps) != 0 {
			splts := strings.SplitN(name, ".", 2)
			ch, chf := C.ChildComps[splts[0]]
			if chf {
				return ch.GetVar(splts[1])
			}
		}
		return C.Session.GetVar(name)
	} else {
		val, found := C.GoVars[name]
		if !found {
			return ""
		}
		return val
	}

}

func (C *Component) SetVars(vars map[string]string) {
	for k, v := range vars {
		C.GoVars[k] = v
	}
}

func (C *ComponentSrc) SetFunc(name string, val func(*Component) string) *ComponentSrc {
	C.GoFuncs[name] = val
	return C
}

func (C *Component) CallFunc(name string) string {

	Pathed := strings.Contains(name, ".")

	if Pathed {
		if C.ChildComps != nil && len(C.ChildComps) != 0 {
			splts := strings.SplitN(name, ".", 2)
			ch, chf := C.ChildComps[splts[0]]
			if chf {
				return ch.CallFunc(splts[1])
			}
		}
		return C.Session.CallFunc(name)
	} else {
		fn, found := C.Src.GoFuncs[name]
		if !found {
			val, varfound := C.GoVars[name]
			if !varfound {
				return ""
			}
			return val
		}

		return fn(C)
	}

}

// func(sender, para) Response
func (C *ComponentSrc) SetEvent(id string, val func(*Component, string, string) EventResponse) *ComponentSrc {
	C.GoEvents[id] = val
	return C
}

func (C *Component) CallEvent(id string, sender string, para string) EventResponse {

	if C.Src == nil {
		return EventResponse{
			Update: DOMEffect{},
		}
	}

	ev, found := C.Src.GoEvents[id]
	if !found {
		return EventResponse{
			Update: DOMEffect{},
		}
	}

	if len(sender) == 0 {
		sender = "nosender"
	}

	if len(para) == 0 {
		para = "null"
	}

	return ev(C, sender, para)
}

// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //
// -------------------------------------------------------------------------- //

type DOMEffect struct {
	Rerender      map[string]string // [root.compid1.compid2]content
	GoVarChanges  map[string]string // [root.govarname]value
	GoFuncChanges map[string]string // [root.gofuncname]value
}

func (F *DOMEffect) New(C Component) {

	F.Rerender = map[string]string{}
	F.GoVarChanges = map[string]string{}
	F.GoFuncChanges = map[string]string{}

}

func (F *DOMEffect) AddRerender(C Component) {
	F.Rerender[C.Path+"."+C.ID] = C.Render()
}

func (F *DOMEffect) AddVar(C Component, varname string) {
	F.GoFuncChanges[C.Path+"."+C.ID+"."+varname] = C.GetVar(varname)
}

func (F *DOMEffect) AddFunc(C Component, varname string) {
	F.GoFuncChanges[C.Path+"."+C.ID+"."+varname] = C.GetVar(varname)
}
