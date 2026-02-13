package treesitter

import (
	"path/filepath"
	"strings"

	engine "github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
	sitter "github.com/smacker/go-tree-sitter"
)

type Visitor struct {
	ad    LangAdapter
	file  *pb.File
	ns    *pb.StmtNamespace
	lines []string

	classStk []*pb.StmtClass
	funcStk  []*pb.StmtFunction
}

func (v *Visitor) curStmts() *pb.Stmts {
	if f := v.curFunc(); f != nil {
		return f.Stmts
	}
	if c := v.curClass(); c != nil {
		return c.Stmts
	}
	return v.file.Stmts
}

func NewVisitor(ad LangAdapter, path string, src []byte) *Visitor {
	lines := strings.Split(string(src), "\n")
	mod := ad.ModuleNameFromPath(filepath.Base(path))

	return &Visitor{
		ad:    ad,
		file:  &pb.File{Path: path, ProgrammingLanguage: "", Stmts: engine.FactoryStmts(), LinesOfCode: &pb.LinesOfCode{LinesOfCode: int32(len(lines))}},
		ns:    &pb.StmtNamespace{Name: &pb.Name{Short: mod, Qualified: mod}, Stmts: engine.FactoryStmts(), LinesOfCode: &pb.LinesOfCode{}},
		lines: lines,
	}
}

func (v *Visitor) Result() *pb.File {
	if len(v.file.Stmts.StmtNamespace) == 0 {
		v.file.Stmts.StmtNamespace = append(v.file.Stmts.StmtNamespace, v.ns)
	}

	// allow adapter to provide a better comment count
	if cc, ok := v.ad.(interface{ CountComments([]string, int, int) int }); ok {
		newC := int32(cc.CountComments(v.lines, 1, len(v.lines)))
		v.file.LinesOfCode.CommentLinesOfCode = newC
		// also recompute LLOC and NCLOC at file-level using blank lines and updated CLOC
		blank := 0
		for _, ln := range v.lines {
			if strings.TrimSpace(ln) == "" {
				blank++
			}
		}
		loc := int(v.file.LinesOfCode.LinesOfCode)
		cloc := int(v.file.LinesOfCode.CommentLinesOfCode)
		offset := 2
		if tun, ok := v.ad.(interface{ FileLlocOffset() int }); ok {
			offset = tun.FileLlocOffset()
		}
		lloc := loc - (cloc + blank + offset)
		if lloc < 0 {
			lloc = 0
		}
		ncloc := loc - cloc
		v.file.LinesOfCode.LogicalLinesOfCode = int32(lloc)
		v.file.LinesOfCode.NonCommentLinesOfCode = int32(ncloc)
	}

	return v.file
}

func (v *Visitor) pushClass(c *pb.StmtClass) {
	v.classStk = append(v.classStk, c)
}

func (v *Visitor) popClass() {
	v.classStk = v.classStk[:len(v.classStk)-1]
}

func (v *Visitor) curClass() *pb.StmtClass {
	if len(v.classStk) == 0 {
		return nil
	}
	return v.classStk[len(v.classStk)-1]
}

func (v *Visitor) pushFunc(f *pb.StmtFunction) {
	v.funcStk = append(v.funcStk, f)
}

func (v *Visitor) popFunc() {
	v.funcStk = v.funcStk[:len(v.funcStk)-1]
}

func (v *Visitor) curFunc() *pb.StmtFunction {
	if len(v.funcStk) == 0 {
		return nil
	}
	return v.funcStk[len(v.funcStk)-1]
}

func (v *Visitor) attachClass(c *pb.StmtClass) {
	v.ns.Stmts.StmtClass = append(v.ns.Stmts.StmtClass, c)
	if f := v.curFunc(); f != nil {
		f.Stmts.StmtClass = append(f.Stmts.StmtClass, c)
		return
	}
	if pc := v.curClass(); pc != nil {
		pc.Stmts.StmtClass = append(pc.Stmts.StmtClass, c)
		return
	}
	v.file.Stmts.StmtClass = append(v.file.Stmts.StmtClass, c)
}

func (v *Visitor) attachFunction(fn *pb.StmtFunction) {
	v.ns.Stmts.StmtFunction = append(v.ns.Stmts.StmtFunction, fn)
	if f := v.curFunc(); f != nil {
		f.Stmts.StmtFunction = append(f.Stmts.StmtFunction, fn)
		return
	}
	if pc := v.curClass(); pc != nil {
		pc.Stmts.StmtFunction = append(pc.Stmts.StmtFunction, fn)
		return
	}
	v.file.Stmts.StmtFunction = append(v.file.Stmts.StmtFunction, fn)
}

// Optional interface support
// An adapter can implement this to let Visitor create StmtInterface nodes.
type InterfaceAware interface {
	IsInterface(*sitter.Node) bool
}

func (v *Visitor) Visit(node *sitter.Node) {
	switch {
	case v.ad.IsModule(node):
		for i := 0; i < int(node.ChildCount()); i++ {
			v.Visit(node.Child(i))
		}
		return
	case func() bool {
		if ia, ok := v.ad.(InterfaceAware); ok {
			return ia.IsInterface(node)
		}
		return false
	}():
		name := v.ad.NodeName(node)
		qualified := name
		if v.ns != nil && v.ns.Name != nil {
			ns := v.ns.Name.Qualified
			if ns != "" {
				qualified = ns + "\\" + name
			}
		}
		itf := &pb.StmtInterface{
			Name:  &pb.Name{Short: name, Qualified: qualified},
			Stmts: engine.FactoryStmts(),
		}
		body := v.ad.NodeBody(node)
		// attach to namespace and file
		v.ns.Stmts.StmtInterface = append(v.ns.Stmts.StmtInterface, itf)
		v.file.Stmts.StmtInterface = append(v.file.Stmts.StmtInterface, itf)
		// visit body
		v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
		return

	case v.ad.IsClass(node):
		name := v.ad.NodeName(node)
		qualified := name
		// qualify with namespace if provided (PHP namespaces, even single segment)
		if v.ns != nil && v.ns.Name != nil {
			ns := v.ns.Name.Qualified
			if ns != "" {
				qualified = ns + "\\" + name
			}
		}
		c := &pb.StmtClass{
			Name:        &pb.Name{Short: name, Qualified: qualified},
			Stmts:       engine.FactoryStmts(),
			LinesOfCode: &pb.LinesOfCode{},
		}
		body := v.ad.NodeBody(node)
		start := int(node.StartPoint().Row) + 1
		end := start
		if body != nil {
			// For class LOC, count from the class declaration line up to the closing brace line inclusively.
			// body.EndPoint().Row points at the '}' line; do not add +1 here to avoid counting the next line.
			end = max(start, int(body.EndPoint().Row))
		}
		c.LinesOfCode = engine.GetLocPositionFromSource(v.lines, start, end)
		// If adapter can count comments precisely (e.g., PHP docblocks), override class CLOC using adapter for class span
		if cc, ok := v.ad.(interface{ CountComments([]string, int, int) int }); ok {
			newC := int32(cc.CountComments(v.lines, start, end))
			c.LinesOfCode.CommentLinesOfCode = newC
		}

		// Pre-initialize class-level CLOC from class body to preserve expected semantics in tests
		if c.Stmts == nil {
			c.Stmts = engine.FactoryStmts()
		}
		if c.Stmts.Analyze == nil {
			c.Stmts.Analyze = &pb.Analyze{}
		}
		if c.Stmts.Analyze.Volume == nil {
			c.Stmts.Analyze.Volume = &pb.Volume{}
		}
		cl := c.LinesOfCode.CommentLinesOfCode
		c.Stmts.Analyze.Volume.Cloc = &cl

		v.attachClass(c)
		// Attach any class-level externals provided by adapter
		if items := v.ad.Imports(node); len(items) > 0 {
			for _, it := range items {
				name := it.Name // leave empty for plain module imports (Python expectation)
				from := ""
				if f := v.curFunc(); f != nil && f.Name != nil {
					from = f.Name.Qualified
					if from == "" {
						from = f.Name.Short
					}
				} else if c != nil && c.Name != nil {
					from = c.Name.Qualified
					if from == "" {
						from = c.Name.Short
					}
				} else if v.ns != nil && v.ns.Name != nil {
					from = v.ns.Name.Qualified
					if from == "" {
						from = v.ns.Name.Short
					}
				}
				dep := &pb.StmtExternalDependency{ClassName: name, Namespace: it.Module, From: from}
				c.Stmts.StmtExternalDependencies = append(c.Stmts.StmtExternalDependencies, dep)
			}
		}
		// If adapter can list direct class operands (e.g., PHP properties), attach them
		if va, ok := v.ad.(interface{ ClassDirectOperands(*sitter.Node) []string }); ok {
			for _, p := range va.ClassDirectOperands(node) {
				c.Operands = append(c.Operands, &pb.StmtOperand{Name: p})
			}
		}

		v.pushClass(c)
		v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
		v.popClass()
		return

	case v.ad.IsFunction(node):
		name := v.ad.NodeName(node)
		qualified := name
		if cls := v.curClass(); cls != nil {
			qualified = v.ad.AttachQualified(cls.Name.Qualified, name)
		}

		fn := &pb.StmtFunction{
			Name:        &pb.Name{Short: name, Qualified: qualified},
			Stmts:       engine.FactoryStmts(),
			LinesOfCode: &pb.LinesOfCode{},
		}
		if params := v.ad.NodeParams(node); params != nil {
			v.ad.EachParamIdent(params, func(id string) {
				fn.Parameters = append(fn.Parameters, &pb.StmtParameter{Name: id})
			})
		}
		body := v.ad.NodeBody(node)
		nodeStart := int(node.StartPoint().Row) + 1
		nodeEnd := int(node.EndPoint().Row) + 1
		locStart := nodeStart
		locEnd := nodeEnd
		if body != nil {
			locStart = int(body.StartPoint().Row) + 1
			locEnd = int(body.EndPoint().Row) + 1
		}
		fn.LinesOfCode = engine.GetLocPositionFromSource(v.lines, locStart, locEnd)

		// allow adapter to provide a better comment count
		if cc, ok := v.ad.(interface{ CountComments([]string, int, int) int }); ok {
			cs := int(node.StartPoint().Row) + 1
			ce := int(node.EndPoint().Row) + 1
			newC := int32(cc.CountComments(v.lines, cs, ce))
			fn.LinesOfCode.CommentLinesOfCode = newC
		}

		v.attachFunction(fn)
		v.pushFunc(fn)
		v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
		// optional: extract operators/operands from source per adapter
		if va, ok := v.ad.(interface {
			ExtractOperatorsOperands(src []byte, startLine, endLine int) (ops []string, operands []string)
		}); ok {
			ops, opr := va.ExtractOperatorsOperands([]byte(strings.Join(v.lines, "\n")), nodeStart, nodeEnd)
			for _, o := range ops {
				fn.Operators = append(fn.Operators, &pb.StmtOperator{Name: o})
			}
			for _, p := range opr {
				fn.Operands = append(fn.Operands, &pb.StmtOperand{Name: p})
			}
		}
		// optional: extract method calls (e.g., this.foo, parent.bar) per adapter
		if mc, ok := v.ad.(interface {
			ExtractMethodCalls(src []byte, startLine, endLine int) []string
		}); ok {
			calls := mc.ExtractMethodCalls([]byte(strings.Join(v.lines, "\n")), nodeStart, nodeEnd)
			for _, m := range calls {
				fn.MethodCalls = append(fn.MethodCalls, &pb.StmtMethodCall{Name: m})
			}
		}
		v.popFunc()
		return
	}

	// Imports and externals
	if items := v.ad.Imports(node); len(items) > 0 {
		st := v.curStmts()
		for _, it := range items {
			name := it.Name // keep empty for plain imports
			from := ""
			if f := v.curFunc(); f != nil && f.Name != nil {
				from = f.Name.Qualified
				if from == "" {
					from = f.Name.Short
				}
			} else if c := v.curClass(); c != nil && c.Name != nil {
				from = c.Name.Qualified
				if from == "" {
					from = c.Name.Short
				}
			} else if v.ns != nil && v.ns.Name != nil {
				from = v.ns.Name.Qualified
				if from == "" {
					from = v.ns.Name.Short
				}
			}
			dep := &pb.StmtExternalDependency{
				ClassName:    name,
				FunctionName: "",
				Namespace:    it.Module,
				From:         from,
			}
			// attach to class scope when inside a class to satisfy PHP tests
			if c := v.curClass(); c != nil {
				c.Stmts.StmtExternalDependencies = append(c.Stmts.StmtExternalDependencies, dep)
			}
			st.StmtExternalDependencies = append(st.StmtExternalDependencies, dep)
			v.ns.Stmts.StmtExternalDependencies = append(v.ns.Stmts.StmtExternalDependencies, dep)
		}
	}

	// Decisions
	if kind, body := v.ad.Decision(node); kind != DecNone {
		st := v.curStmts()
		switch kind {
		case DecIf:
			ifn := &pb.StmtDecisionIf{Stmts: engine.FactoryStmts()}
			st.StmtDecisionIf = append(st.StmtDecisionIf, ifn)
			// Visit the if body
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })

			// Iterate over siblings elif/else of if_statement
			for i := 0; i < int(node.ChildCount()); i++ {
				ch := node.Child(i)
				k2, b2 := v.ad.Decision(ch)
				switch k2 {
				case DecElif:
					// If adapter wants elseif to be treated as an if (PHP), record only as if; otherwise record as elseif
					if x, ok := v.ad.(interface{ CountElseIfAsIf() bool }); ok && x.CountElseIfAsIf() {
						st.StmtDecisionIf = append(st.StmtDecisionIf, &pb.StmtDecisionIf{Stmts: engine.FactoryStmts()})
					} else {
						st.StmtDecisionElseIf = append(st.StmtDecisionElseIf, &pb.StmtDecisionElseIf{Stmts: engine.FactoryStmts()})
					}
					v.ad.EachChildBody(b2, func(cci *sitter.Node) { v.Visit(cci) })
				case DecElse:
					el := &pb.StmtDecisionElse{Stmts: engine.FactoryStmts()}
					st.StmtDecisionElse = append(st.StmtDecisionElse, el)
					v.ad.EachChildBody(b2, func(cci *sitter.Node) { v.Visit(cci) })
				}
			}
			return

		case DecElif:
			// If adapter wants elseif as if (PHP), record only as if; else record as elseif
			if x, ok := v.ad.(interface{ CountElseIfAsIf() bool }); ok && x.CountElseIfAsIf() {
				st.StmtDecisionIf = append(st.StmtDecisionIf, &pb.StmtDecisionIf{Stmts: engine.FactoryStmts()})
			} else {
				st.StmtDecisionElseIf = append(st.StmtDecisionElseIf, &pb.StmtDecisionElseIf{Stmts: engine.FactoryStmts()})
			}
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecElse:
			el := &pb.StmtDecisionElse{Stmts: engine.FactoryStmts()}
			st.StmtDecisionElse = append(st.StmtDecisionElse, el)
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecLoop:
			lp := &pb.StmtLoop{Stmts: engine.FactoryStmts()}
			st.StmtLoop = append(st.StmtLoop, lp)
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecSwitch:
			sw := &pb.StmtDecisionSwitch{Stmts: engine.FactoryStmts()}
			st.StmtDecisionSwitch = append(st.StmtDecisionSwitch, sw)
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecCase:
			cs := &pb.StmtDecisionCase{Stmts: engine.FactoryStmts()}
			st.StmtDecisionCase = append(st.StmtDecisionCase, cs)
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return
		}
	}

	// Fallback
	for i := 0; i < int(node.ChildCount()); i++ {
		v.Visit(node.Child(i))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
