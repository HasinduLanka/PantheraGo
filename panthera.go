package main

import (
	"regexp"
	"sort"
	"strings"
)

var RegexGoCompBegin *regexp.Regexp = regexp.MustCompile(`< *gocomp +compinst *= *"(.*?)" +compsrc *= *"(.*?)"(.*?)>`)
var RegexGoCompEnd *regexp.Regexp = regexp.MustCompile(`< *?/gocomp *?>`)
var RegexGo *regexp.Regexp = regexp.MustCompile(`(?s)< *go +(.*?)>(.*?)< */go *>`)
var RegexVar *regexp.Regexp = regexp.MustCompile(`var *= *"(.*?)".*?>`)
var RegexFunc *regexp.Regexp = regexp.MustCompile(`func *= *"(.*?)"(.*?)>`)
var RegexArgs *regexp.Regexp = regexp.MustCompile(` +(.*) *= *"(.*?)"`)

var NewLineRemover *strings.Replacer = strings.NewReplacer("\n", "")

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
	CE := RegexGoCompEnd.FindAllStringSubmatchIndex(HTML, 4096)

	L := len(CB)
	if L != len(CE) {
		return "Component Begin/End Pairs malformed"
	}

	type BoolIntArr struct {
		IsBegin bool
		Arr     []int
	}

	CompIndexes := make(map[int]BoolIntArr, L*2)
	CompIndexArr := make([]int, L*2)
	iCompIndexArr := 0

	for _, v := range CB {
		CompIndexes[v[0]] = BoolIntArr{IsBegin: true, Arr: v}
		CompIndexArr[iCompIndexArr] = v[0]
		iCompIndexArr++
	}
	for _, v := range CE {
		CompIndexes[v[0]] = BoolIntArr{IsBegin: false, Arr: v}
		CompIndexArr[iCompIndexArr] = v[0]
		iCompIndexArr++
	}

	sort.Ints(CompIndexArr)
	depth := 0

	Comps := map[string]*Component{}
	CompsBegins := map[string][]int{}
	CompList := []string{}

	for i := 0; i < L*2; i++ {
		ci := CompIndexArr[i]
		CI := CompIndexes[ci]

		if CI.IsBegin {

			if depth == 0 {
				CompID := HTML[CI.Arr[2]:CI.Arr[3]]

				var C *Component

				OldC, Old := Parent.ChildComps[CompID]
				if Old {
					if OldC.Src == nil {
						C = OldC
					} else {
						Comps[CompID] = OldC
						CompsBegins[CompID] = CI.Arr[0:2]
						CompList = append(CompList, CompID)
						depth++
						continue
					}
				} else {
					C = &Component{Path: Parent.Path + "." + Parent.ID,
						ID:         CompID,
						Session:    S,
						Parent:     Parent,
						ChildComps: map[string]*Component{},
						GoVars:     map[string]string{}}
				}

				CompSrcs := HTML[CI.Arr[4]:CI.Arr[5]]
				CompSrc := S.ComponentSrcProvider(CompSrcs)
				if CompSrc == nil {
					println("Component source not found ID:" + CompID + " Src:" + CompSrcs)
				}

				C.Src = CompSrc
				Comps[CompID] = C
				CompsBegins[CompID] = CI.Arr[0:2]
				CompList = append(CompList, CompID)

			}
			depth++
		} else {
			depth--
		}

	}

	// CompIndexesJ, JErr := json.Marshal(Comps)
	// PrintError(JErr)
	// return string(CBJ) + " <br> " + string(CEJ) + " <br> -- <br> " + string(CompIndexesJ)

	Parent.ChildComps = Comps

	var sb strings.Builder
	LastI := 0
	for _, v := range CompList {
		CompE := CompsBegins[v][1]
		if LastI < CompE { // Double IDs
			sb.WriteString(HTML[LastI:CompE] + "\n")
			C := Comps[v]
			sb.WriteString(C.Render())
			sb.WriteString("\n</gocomp>\n")
			LastI = CompE
		}
	}

	return sb.String()

}

func (S *Session) RenderComp(gocomp string) string {
	return ""
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
	return RegexGo.ReplaceAllStringFunc(C.Src.Provider(), C.renderTag)

}

func (C *Component) RenderReusable() string {
	if C.Src == nil {
		return "No source for reusable component " + C.ID
	}
	return RegexGo.ReplaceAllStringFunc(C.Src.Provider(), C.renderReusableTag)
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
	return `<span govar="` + T + `">` + C.GetVar(T) + `</span> `
}

func (C *Component) renderFunc(T string) string {
	return `<span gofunc="` + T + `">` + C.CallFunc(T) + `</span> `
}

func (C *Component) SetVar(name string, val string) {
	C.GoVars[name] = val
}

func (C *Component) GetVar(name string) string {
	val, found := C.GoVars[name]
	if !found {
		return ""
	}
	return val
}

func (C *ComponentSrc) SetFunc(name string, val func(*Component) string) {
	C.GoFuncs[name] = val
}

func (C *Component) CallFunc(name string) string {
	fn, found := C.Src.GoFuncs[name]
	if !found {
		return ""
	}

	return fn(C)

}

// func(sender, para) Response
func (C *ComponentSrc) SetEvent(id string, val func(*Component, string, string) EventResponse) {
	C.GoEvents[id] = val
}

func (C *Component) CallEvent(id string, sender string, para string) EventResponse {
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
