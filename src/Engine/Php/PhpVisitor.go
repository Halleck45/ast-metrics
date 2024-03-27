package Php

import (
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/halleck45/ast-metrics/src/Engine"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type PhpVisitor struct {
	*visitor.Null
	file             *pb.File
	linesOfFile      []string
	currentClass     *pb.StmtClass
	currentInterface *pb.StmtInterface
	currentNamespace *pb.StmtNamespace
	currentMethod    *pb.StmtFunction
	currentStmts     *pb.Stmts
}

func (v *PhpVisitor) nameObject(name string) string {
	qualified := ""
	if v.currentNamespace != nil {
		qualified = v.currentNamespace.Name.Qualified
	}

	return qualified + name
}

func (v *PhpVisitor) nameMethod(name string) string {
	qualifiedName := name
	if v.currentClass != nil {
		qualifiedName = v.currentClass.Name.Qualified + "::" + name
	} else if v.currentInterface != nil {
		qualifiedName = v.currentInterface.Name.Qualified + "::" + name
	}

	return qualifiedName
}

func (v *PhpVisitor) StmtClass(node *ast.StmtClass) {

	name := "@anonymous"
	if node.Name != nil {
		name = string(node.Name.(*ast.Identifier).Value)
	}

	class := &pb.StmtClass{
		Name: &pb.Name{
			Short:     name,
			Qualified: v.nameObject(name),
		},
	}
	class.Stmts = Engine.FactoryStmts()
	class.LinesOfCode = &pb.LinesOfCode{}
	v.file.Stmts.StmtClass = append(v.file.Stmts.StmtClass, class)
	v.currentClass = class
	v.currentStmts = class.Stmts
}

// ----------------
// Classes
// ----------------
func (v *PhpVisitor) StmtInterface(node *ast.StmtInterface) {

	name := "@anonymous"
	if node.Name != nil {
		name = string(node.Name.(*ast.Identifier).Value)

	}

	class := &pb.StmtInterface{
		Name: &pb.Name{
			Short:     name,
			Qualified: v.nameObject(name),
		},
	}
	class.Stmts = Engine.FactoryStmts()
	v.file.Stmts.StmtInterface = append(v.file.Stmts.StmtInterface, class)
	v.currentInterface = class
	v.currentStmts = class.Stmts
}

func (v *PhpVisitor) StmtTrait(node *ast.StmtTrait) {

	name := "@anonymous"
	if node.Name != nil {
		name = string(node.Name.(*ast.Identifier).Value)

	}

	class := &pb.StmtClass{
		Name: &pb.Name{
			Short:     name,
			Qualified: v.nameObject(name),
		},
	}
	class.LinesOfCode = &pb.LinesOfCode{}
	class.Stmts = Engine.FactoryStmts()
	v.file.Stmts.StmtClass = append(v.file.Stmts.StmtClass, class)
	v.currentClass = class
	v.currentStmts = class.Stmts
}

func (v *PhpVisitor) StmtEnum(node *ast.StmtEnum) {

	name := "@anonymous"
	if node.Name != nil {
		name = string(node.Name.(*ast.Identifier).Value)

	}

	class := &pb.StmtClass{
		Name: &pb.Name{
			Short:     name,
			Qualified: v.nameObject(name),
		},
	}
	class.LinesOfCode = &pb.LinesOfCode{}
	class.Stmts = Engine.FactoryStmts()
	v.file.Stmts.StmtClass = append(v.file.Stmts.StmtClass, class)
	v.currentClass = class
	v.currentStmts = class.Stmts
}

// ----------------
// Functions
// ----------------
func (v *PhpVisitor) StmtClassMethod(node *ast.StmtClassMethod) {

	if v.currentClass == nil && v.currentInterface == nil {
		// should not happen
		panic("currentClass is nil for StmtClassMethod" + string(node.Name.(*ast.Identifier).Value) + " in file " + v.file.Path)
	}

	name := string(node.Name.(*ast.Identifier).Value)
	method := &pb.StmtFunction{
		Name: &pb.Name{
			Short:     name,
			Qualified: v.nameMethod(name),
		},
	}
	method.Stmts = Engine.FactoryStmts()

	if v.currentClass == nil {
		// should not happen
		return
	}

	// Lines of code
	start := node.GetPosition().StartLine
	end := node.GetPosition().EndLine
	loc := Engine.GetLocPositionFromSource(v.linesOfFile, start, end)
	method.LinesOfCode = loc
	v.findPhpDocBlock(start, end, method.LinesOfCode)

	// Add to class
	v.currentClass.LinesOfCode.LinesOfCode += loc.LinesOfCode
	v.currentClass.LinesOfCode.LogicalLinesOfCode += loc.LogicalLinesOfCode
	v.currentClass.LinesOfCode.CommentLinesOfCode += loc.CommentLinesOfCode

	v.currentClass.Stmts.StmtFunction = append(v.currentClass.Stmts.StmtFunction, method)
	v.currentStmts = method.Stmts
	v.currentMethod = method
}

func (v *PhpVisitor) findPhpDocBlock(start int, end int, linesOfCode *pb.LinesOfCode) {

	// iterate over previous lines of code to find docblock
	endDocBlock := 0
	startDocBlock := 0
	docBlockEndFound := false
	docBlockStartFound := false
	for i := start - 1; i >= 0; i-- {
		line := v.linesOfFile[i]

		// remove leading spaces
		line = strings.TrimSpace(line)

		if len(line) < 2 {
			continue
		}

		if line[0] == '/' && line[1] == '/' {
			continue
		}

		if line[0] == '/' && line[1] == '*' {
			docBlockStartFound = true
			startDocBlock = i
			break
		}

		if line[0] == '*' && line[1] == '/' {
			endDocBlock = i
			docBlockEndFound = true
		}
	}

	if docBlockEndFound && docBlockStartFound && endDocBlock > startDocBlock {
		nbLinesInDocBlock := endDocBlock - startDocBlock + 1
		linesOfCode.CommentLinesOfCode += int32(nbLinesInDocBlock)
	}
}

func (v *PhpVisitor) StmtFunction(node *ast.StmtFunction) {
	name := string(node.Name.(*ast.Identifier).Value)

	method := &pb.StmtFunction{
		Name: &pb.Name{
			Short:     name,
			Qualified: name,
		},
	}
	method.Stmts = Engine.FactoryStmts()

	// Lines of code
	start := node.GetPosition().StartLine
	end := node.GetPosition().EndLine
	loc := Engine.GetLocPositionFromSource(v.linesOfFile, start, end)
	method.LinesOfCode = loc
	v.findPhpDocBlock(start, end, method.LinesOfCode)

	v.file.Stmts.StmtFunction = append(v.file.Stmts.StmtFunction, method)
	v.currentStmts = method.Stmts
	v.currentMethod = method
}

func (v *PhpVisitor) StmtNamespace(node *ast.StmtNamespace) {

	name := ""
	if node.Name != nil {
		// if namespace has no name, it is global namespace
		parts := node.Name.(*ast.Name).Parts
		for _, part := range parts {
			name += string(part.(*ast.NamePart).Value) + "\\"
		}
	}

	namespace := &pb.StmtNamespace{
		Name: &pb.Name{
			Short:     name,
			Qualified: name,
		},
	}
	namespace.Stmts = Engine.FactoryStmts()

	v.file.Stmts.StmtNamespace = append(v.file.Stmts.StmtNamespace, namespace)
	v.currentNamespace = namespace
}

// ----------------
// Loops
// ----------------
func (v *PhpVisitor) StmtFor(node *ast.StmtFor) {
	if v.currentStmts == nil {
		return
	}
	v.currentStmts.StmtLoop = append(v.currentStmts.StmtLoop, &pb.StmtLoop{})
}

func (v *PhpVisitor) StmtForeach(node *ast.StmtForeach) {
	if v.currentStmts == nil {
		return
	}
	v.currentStmts.StmtLoop = append(v.currentStmts.StmtLoop, &pb.StmtLoop{})
}

func (v *PhpVisitor) StmtWhile(node *ast.StmtWhile) {
	if v.currentStmts == nil {
		return
	}
	v.currentStmts.StmtLoop = append(v.currentStmts.StmtLoop, &pb.StmtLoop{})
}

func (v *PhpVisitor) StmtDo(node *ast.StmtDo) {
	if v.currentStmts == nil {
		return
	}
	v.currentStmts.StmtLoop = append(v.currentStmts.StmtLoop, &pb.StmtLoop{})
}

// ----------------
// Conditions
// ----------------
func (v *PhpVisitor) StmtIf(node *ast.StmtIf) {
	if v.currentStmts == nil {
		return
	}
	v.currentStmts.StmtDecisionIf = append(v.currentStmts.StmtDecisionIf, &pb.StmtDecisionIf{})
}

func (v *PhpVisitor) StmtElseIf(node *ast.StmtElseIf) {
	if v.currentStmts == nil {
		return
	}
	v.currentStmts.StmtDecisionIf = append(v.currentStmts.StmtDecisionIf, &pb.StmtDecisionIf{})
}

func (v *PhpVisitor) EnumCase(node *ast.EnumCase) {
	if v.currentStmts == nil {
		return
	}
	v.currentStmts.StmtDecisionSwitch = append(v.currentStmts.StmtDecisionSwitch, &pb.StmtDecisionSwitch{})
}

func (v *PhpVisitor) StmtElse(node *ast.StmtElse) {
}

// ----------------
// Operands
// ----------------
func (v *PhpVisitor) ExprVariable(node *ast.ExprVariable) {
	if v.currentMethod == nil {
		return
	}

	// Ensure identifier
	_, ok := node.Name.(*ast.Identifier)
	if !ok {
		// dynamic variable like $$foo ok ${"foo"} are not supported
		return
	}

	name := string(node.Name.(*ast.Identifier).Value)
	if name == "$this" || name == "self" {
		return
	}
	operand := &pb.StmtOperand{Name: name}
	v.currentMethod.Operands = append(v.currentMethod.Operands, operand)
}

func (v *PhpVisitor) ExprPropertyFetch(node *ast.ExprPropertyFetch) {
	if v.currentMethod == nil {
		return
	}

	// if method call, do not add it as operand
	// ast.Vertex is *ast.ExprMethodCall, not *ast.ExprVariable
	_, ok := node.Var.(*ast.ExprVariable)
	if !ok {
		return
	}
	// Ensure identifier
	_, ok = node.Var.(*ast.ExprVariable).Name.(*ast.Identifier)
	if !ok {
		return
	}

	name := string(node.Var.(*ast.ExprVariable).Name.(*ast.Identifier).Value)

	// ast.Vertex is *ast.ExprBrackets, not *ast.Identifier
	// $this->{foo}
	_, ok = node.Prop.(*ast.Identifier)
	if !ok {
		return
	}
	name += "->" + string(node.Prop.(*ast.Identifier).Value)

	operand := &pb.StmtOperand{Name: name}
	v.currentMethod.Operands = append(v.currentMethod.Operands, operand)
}

// ----------------
// Operators
// ----------------
func (v *PhpVisitor) ExprAssign(node *ast.ExprAssign) {
}

func (v *PhpVisitor) ExprAssignReference(node *ast.ExprAssignReference) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "="})
}

func (v *PhpVisitor) ExprAssignBitwiseAnd(node *ast.ExprAssignBitwiseAnd) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "&="})
}

func (v *PhpVisitor) ExprAssignBitwiseOr(node *ast.ExprAssignBitwiseOr) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "|="})
}

func (v *PhpVisitor) ExprAssignBitwiseXor(node *ast.ExprAssignBitwiseXor) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "^="})
}

func (v *PhpVisitor) ExprAssignCoalesce(node *ast.ExprAssignCoalesce) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "??="})
}

func (v *PhpVisitor) ExprAssignConcat(node *ast.ExprAssignConcat) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: ".="})
}

func (v *PhpVisitor) ExprAssignDiv(node *ast.ExprAssignDiv) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "/="})
}

func (v *PhpVisitor) ExprAssignMinus(node *ast.ExprAssignMinus) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "-="})
}

func (v *PhpVisitor) ExprAssignMod(node *ast.ExprAssignMod) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "%="})
}

func (v *PhpVisitor) ExprAssignMul(node *ast.ExprAssignMul) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "*="})
}

func (v *PhpVisitor) ExprAssignPlus(node *ast.ExprAssignPlus) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "+="})
}

func (v *PhpVisitor) ExprAssignPow(node *ast.ExprAssignPow) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "**="})
}

func (v *PhpVisitor) ExprAssignShiftLeft(node *ast.ExprAssignShiftLeft) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "<<="})
}

func (v *PhpVisitor) ExprAssignShiftRight(node *ast.ExprAssignShiftRight) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: ">>="})
}

func (v *PhpVisitor) ExprBinaryBitwiseAnd(node *ast.ExprBinaryBitwiseAnd) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "&"})
}

func (v *PhpVisitor) ExprBinaryBitwiseOr(node *ast.ExprBinaryBitwiseOr) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "|"})
}

func (v *PhpVisitor) ExprBinaryBitwiseXor(node *ast.ExprBinaryBitwiseXor) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "^"})
}

func (v *PhpVisitor) ExprBinaryBooleanAnd(node *ast.ExprBinaryBooleanAnd) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "&&"})
}

func (v *PhpVisitor) ExprBinaryBooleanOr(node *ast.ExprBinaryBooleanOr) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "||"})
}

func (v *PhpVisitor) ExprBinaryCoalesce(node *ast.ExprBinaryCoalesce) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "??="})
}

func (v *PhpVisitor) ExprBinaryConcat(node *ast.ExprBinaryConcat) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "."})
}

func (v *PhpVisitor) ExprBinaryDiv(node *ast.ExprBinaryDiv) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "/"})
}

func (v *PhpVisitor) ExprBinaryEqual(node *ast.ExprBinaryEqual) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "=="})
}

func (v *PhpVisitor) ExprBinaryGreater(node *ast.ExprBinaryGreater) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: ">"})
}

func (v *PhpVisitor) ExprBinaryGreaterOrEqual(node *ast.ExprBinaryGreaterOrEqual) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: ">="})
}

func (v *PhpVisitor) ExprBinaryIdentical(node *ast.ExprBinaryIdentical) {
}

func (v *PhpVisitor) ExprBinaryLogicalAnd(node *ast.ExprBinaryLogicalAnd) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "&&"})
}

func (v *PhpVisitor) ExprBinaryLogicalOr(node *ast.ExprBinaryLogicalOr) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "||"})
}

func (v *PhpVisitor) ExprBinaryLogicalXor(node *ast.ExprBinaryLogicalXor) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "xor"})
}

func (v *PhpVisitor) ExprBinaryMinus(node *ast.ExprBinaryMinus) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "-"})
}

func (v *PhpVisitor) ExprBinaryMod(node *ast.ExprBinaryMod) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "%"})
}

func (v *PhpVisitor) ExprBinaryMul(node *ast.ExprBinaryMul) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "*"})
}

func (v *PhpVisitor) ExprBinaryNotEqual(node *ast.ExprBinaryNotEqual) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "!="})
}

func (v *PhpVisitor) ExprBinaryNotIdentical(node *ast.ExprBinaryNotIdentical) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "!==="})
}

func (v *PhpVisitor) ExprBinaryPlus(node *ast.ExprBinaryPlus) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "+"})
}

func (v *PhpVisitor) ExprBinaryPow(node *ast.ExprBinaryPow) {
	if v.currentMethod == nil {
		return
	}

	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "**"})
}

func (v *PhpVisitor) ExprBinaryShiftLeft(node *ast.ExprBinaryShiftLeft) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "<<="})
}

func (v *PhpVisitor) ExprBinaryShiftRight(node *ast.ExprBinaryShiftRight) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: ">>="})
}

func (v *PhpVisitor) ExprBinarySmaller(node *ast.ExprBinarySmaller) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "<"})
}

func (v *PhpVisitor) ExprBinarySmallerOrEqual(node *ast.ExprBinarySmallerOrEqual) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "<="})
}

func (v *PhpVisitor) ExprBinarySpaceship(node *ast.ExprBinarySpaceship) {
	if v.currentMethod == nil {
		return
	}
	v.currentMethod.Operators = append(v.currentMethod.Operators, &pb.StmtOperator{Name: "<=>"})
}
