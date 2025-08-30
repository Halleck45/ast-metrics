package Treesitter

import (
	"path/filepath"
	"strings"

	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
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
		file:  &pb.File{Path: path, ProgrammingLanguage: "", Stmts: Engine.FactoryStmts(), LinesOfCode: &pb.LinesOfCode{LinesOfCode: int32(len(lines))}},
		ns:    &pb.StmtNamespace{Name: &pb.Name{Short: mod, Qualified: mod}, Stmts: Engine.FactoryStmts(), LinesOfCode: &pb.LinesOfCode{}},
		lines: lines,
	}
}

func (v *Visitor) Result() *pb.File {
	if len(v.file.Stmts.StmtNamespace) == 0 {
		v.file.Stmts.StmtNamespace = append(v.file.Stmts.StmtNamespace, v.ns)
	}
	return v.file
}

func (v *Visitor) pushClass(c *pb.StmtClass) { v.classStk = append(v.classStk, c) }
func (v *Visitor) popClass()                 { v.classStk = v.classStk[:len(v.classStk)-1] }
func (v *Visitor) curClass() *pb.StmtClass {
	if len(v.classStk) == 0 {
		return nil
	}
	return v.classStk[len(v.classStk)-1]
}
func (v *Visitor) pushFunc(f *pb.StmtFunction) { v.funcStk = append(v.funcStk, f) }
func (v *Visitor) popFunc()                    { v.funcStk = v.funcStk[:len(v.funcStk)-1] }
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
			Stmts: Engine.FactoryStmts(),
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
			Stmts:       Engine.FactoryStmts(),
			LinesOfCode: &pb.LinesOfCode{},
		}
		body := v.ad.NodeBody(node)
		start := int(node.StartPoint().Row) + 1
		end := start
		if body != nil {
			end = max(start, int(body.EndPoint().Row)+1)
		}
		c.LinesOfCode = Engine.GetLocPositionFromSource(v.lines, start, end)

		v.attachClass(c)
		// Attach any class-level externals provided by adapter
		if items := v.ad.Imports(node); len(items) > 0 {
			for _, it := range items {
				name := it.Name // leave empty for plain module imports (Python expectation)
				dep := &pb.StmtExternalDependency{ClassName: name, Namespace: it.Module, From: it.Module}
				c.Stmts.StmtExternalDependencies = append(c.Stmts.StmtExternalDependencies, dep)
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
			Stmts:       Engine.FactoryStmts(),
			LinesOfCode: &pb.LinesOfCode{},
		}
		if params := v.ad.NodeParams(node); params != nil {
			v.ad.EachParamIdent(params, func(id string) {
				fn.Parameters = append(fn.Parameters, &pb.StmtParameter{Name: id})
			})
		}
		body := v.ad.NodeBody(node)
		start := int(node.StartPoint().Row) + 1
		end := int(node.EndPoint().Row) + 1
		if body != nil {
			start = int(body.StartPoint().Row) + 1
			end = int(body.EndPoint().Row) + 1
		}
		fn.LinesOfCode = Engine.GetLocPositionFromSource(v.lines, start, end)
		// allow adapter to provide a better comment count
		if cc, ok := v.ad.(interface{ CountComments([]string, int, int) int }); ok {
			cs := int(node.StartPoint().Row) + 1
			ce := int(node.EndPoint().Row) + 1
			fn.LinesOfCode.CommentLinesOfCode = int32(cc.CountComments(v.lines, cs, ce))
		}

		v.attachFunction(fn)
		v.pushFunc(fn)
		v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
		// optional: extract operators/operands from source per adapter
		if va, ok := v.ad.(interface{ ExtractOperatorsOperands(src []byte, startLine, endLine int) (ops []string, operands []string) }); ok {
			ops, opr := va.ExtractOperatorsOperands([]byte(strings.Join(v.lines, "\n")), start, end)
			for _, o := range ops { fn.Operators = append(fn.Operators, &pb.StmtOperator{Name: o}) }
			for _, p := range opr { fn.Operands = append(fn.Operands, &pb.StmtOperand{Name: p}) }
		}
		v.popFunc()
		return
	}

	// Imports and externals
	if items := v.ad.Imports(node); len(items) > 0 {
		st := v.curStmts()
		for _, it := range items {
			name := it.Name // keep empty for plain imports
			dep := &pb.StmtExternalDependency{
				ClassName:    name,
				FunctionName: "",
				Namespace:    it.Module,
				From:         it.Module,
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
			ifn := &pb.StmtDecisionIf{Stmts: Engine.FactoryStmts()}
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
						st.StmtDecisionIf = append(st.StmtDecisionIf, &pb.StmtDecisionIf{Stmts: Engine.FactoryStmts()})
					} else {
						st.StmtDecisionElseIf = append(st.StmtDecisionElseIf, &pb.StmtDecisionElseIf{Stmts: Engine.FactoryStmts()})
					}
					v.ad.EachChildBody(b2, func(cci *sitter.Node) { v.Visit(cci) })
				case DecElse:
					el := &pb.StmtDecisionElse{Stmts: Engine.FactoryStmts()}
					st.StmtDecisionElse = append(st.StmtDecisionElse, el)
					v.ad.EachChildBody(b2, func(cci *sitter.Node) { v.Visit(cci) })
				}
			}
			return

		case DecElif:
			// If adapter wants elseif as if (PHP), record only as if; else record as elseif
			if x, ok := v.ad.(interface{ CountElseIfAsIf() bool }); ok && x.CountElseIfAsIf() {
				st.StmtDecisionIf = append(st.StmtDecisionIf, &pb.StmtDecisionIf{Stmts: Engine.FactoryStmts()})
			} else {
				st.StmtDecisionElseIf = append(st.StmtDecisionElseIf, &pb.StmtDecisionElseIf{Stmts: Engine.FactoryStmts()})
			}
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecElse:
			el := &pb.StmtDecisionElse{Stmts: Engine.FactoryStmts()}
			st.StmtDecisionElse = append(st.StmtDecisionElse, el)
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecLoop:
			lp := &pb.StmtLoop{Stmts: Engine.FactoryStmts()}
			st.StmtLoop = append(st.StmtLoop, lp)
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecSwitch:
			sw := &pb.StmtDecisionSwitch{Stmts: Engine.FactoryStmts()}
			st.StmtDecisionSwitch = append(st.StmtDecisionSwitch, sw)
			v.ad.EachChildBody(body, func(ch *sitter.Node) { v.Visit(ch) })
			return

		case DecCase:
			cs := &pb.StmtDecisionCase{Stmts: Engine.FactoryStmts()}
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
