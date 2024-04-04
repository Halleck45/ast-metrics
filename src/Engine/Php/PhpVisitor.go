package Php

import (
	"strings"
	"unicode/utf8"

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
	aliases          map[string]string
}

func (v *PhpVisitor) nameObject(name string) string {
	qualified := ""
	if v.currentNamespace != nil {
		qualified = v.currentNamespace.Name.Qualified
	}

	return qualified + name
}

func (v *PhpVisitor) FixName(name *pb.Name) *pb.Name {

	// in PHP classname can be a non UTF-8 string
	// @see https://github.com/symfony/symfony/discussions/46477
	if !utf8.ValidString(name.Qualified) {
		name.Qualified = "@non-utf8"
		if !utf8.ValidString(name.Short) {
			name.Short = "@non-utf8"
		}
	}

	return name
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

	class.Name = v.FixName(class.Name)

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

	class.Name = v.FixName(class.Name)

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

	class.Name = v.FixName(class.Name)
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

	class.Name = v.FixName(class.Name)
	class.LinesOfCode = &pb.LinesOfCode{}
	class.Stmts = Engine.FactoryStmts()
	v.file.Stmts.StmtClass = append(v.file.Stmts.StmtClass, class)
	v.currentClass = class
	v.currentStmts = class.Stmts
}

// ----------------
// Functions
// ----------------
func (v *PhpVisitor) extractParams(params []ast.Vertex) []*pb.StmtParameter {
	for _, param := range params {
		paramType := ""
		if param.(*ast.Parameter).Type != nil {

			parts := make([]ast.Vertex, 0)
			switch param.(*ast.Parameter).Type.(type) {
			case *ast.Name:
				// usage of name, like new DateTime()
				parts = param.(*ast.Parameter).Type.(*ast.Name).Parts
			case *ast.NameFullyQualified:
				// usage of fully qualified name, like new \DateTime()
				parts = param.(*ast.Parameter).Type.(*ast.NameFullyQualified).Parts
				parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
			case *ast.Nullable:
				// usage of fully nullable name, like new ?\DateTime()
				expr := param.(*ast.Parameter).Type.(*ast.Nullable).Expr
				switch expr.(type) {
				case *ast.Name:
					// usage of name, like new DateTime()
					parts = expr.(*ast.Name).Parts
				case *ast.NameFullyQualified:
					// usage of fully qualified name, like new \DateTime()
					parts = expr.(*ast.NameFullyQualified).Parts
					parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
				case *ast.Identifier:
					// usage of name, like new DateTime()
					parts = []ast.Vertex{&ast.NamePart{Value: []byte(expr.(*ast.Identifier).Value)}}
				}
			default:
				// Handle unexpected types
			}

			dependency := v.nameDependencyFromParts(parts)
			if dependency != nil {
				if v.currentMethod.Stmts.StmtExternalDependencies == nil {
					v.currentMethod.Stmts.StmtExternalDependencies = make([]*pb.StmtExternalDependency, 0)
				}
				v.currentMethod.Stmts.StmtExternalDependencies = append(v.currentMethod.Stmts.StmtExternalDependencies, dependency)

				// Add it also to the class dependencies
				if v.currentClass != nil {
					if v.currentClass.Stmts.StmtExternalDependencies == nil {
						v.currentClass.Stmts.StmtExternalDependencies = make([]*pb.StmtExternalDependency, 0)
					}
					v.currentClass.Stmts.StmtExternalDependencies = append(v.currentClass.Stmts.StmtExternalDependencies, dependency)
				}
			}

			// Add to the parameter list
			paramName := string(param.(*ast.Parameter).Var.(*ast.ExprVariable).Name.(*ast.Identifier).Value)
			if v.currentMethod.Parameters == nil {
				v.currentMethod.Parameters = make([]*pb.StmtParameter, 0)
			}
			v.currentMethod.Parameters = append(v.currentMethod.Parameters, &pb.StmtParameter{
				Name: paramName,
				Type: paramType,
			})
		}
	}

	return v.currentMethod.Parameters
}
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

	// Extract parameters and add it to the method, including dependencies
	params := node.Params
	v.extractParams(params)

	// return type
	if node.ReturnType != nil {

		var parts []ast.Vertex
		switch node.ReturnType.(type) {
		case *ast.Name:
			parts = node.ReturnType.(*ast.Name).Parts
		case *ast.NameFullyQualified:
			parts = node.ReturnType.(*ast.NameFullyQualified).Parts
			parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
		case *ast.Nullable:
			// usage of fully nullable name, like new ?\DateTime()
			expr := node.ReturnType.(*ast.Nullable).Expr
			switch expr.(type) {
			case *ast.Name:
				// usage of name, like new DateTime()
				parts = expr.(*ast.Name).Parts
			case *ast.NameFullyQualified:
				// usage of fully qualified name, like new \DateTime()
				parts = expr.(*ast.NameFullyQualified).Parts
				parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
			case *ast.Identifier:
				// usage of name, like new DateTime()
				parts = []ast.Vertex{&ast.NamePart{Value: []byte(expr.(*ast.Identifier).Value)}}
			}
		default:
		}

		dependency := v.nameDependencyFromParts(parts)
		if dependency != nil {
			if v.currentMethod.Stmts.StmtExternalDependencies == nil {
				v.currentMethod.Stmts.StmtExternalDependencies = make([]*pb.StmtExternalDependency, 0)
			}
			v.currentMethod.Stmts.StmtExternalDependencies = append(v.currentMethod.Stmts.StmtExternalDependencies, dependency)

			// Add it also to the class dependencies
			if v.currentClass != nil {
				if v.currentClass.Stmts.StmtExternalDependencies == nil {
					v.currentClass.Stmts.StmtExternalDependencies = make([]*pb.StmtExternalDependency, 0)
				}
				v.currentClass.Stmts.StmtExternalDependencies = append(v.currentClass.Stmts.StmtExternalDependencies, dependency)
			}
		}
	}

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

	// Extract parameters and add it to the method, including dependencies
	params := node.Params
	v.extractParams(params)
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

// ----------------
// Use of external code
// ----------------
func (v *PhpVisitor) unalias(dependency *pb.StmtExternalDependency) *pb.StmtExternalDependency {

	// if no alias, return
	if v.aliases == nil || len(v.aliases) == 0 {
		return dependency
	}

	// If the dependency is an alias, replace it with the real name
	if realName, ok := v.aliases[dependency.ClassName]; ok {
		dependency.ClassName = realName
		return dependency
	}

	// if dependency does not start with slash, use current namespace
	if !strings.HasPrefix(dependency.ClassName, "\\") && v.currentNamespace != nil {
		dependency.ClassName = v.nameObject(dependency.ClassName)
	}

	// trim leading slash
	dependency.ClassName = strings.TrimPrefix(dependency.ClassName, "\\")

	return dependency
}

// Name of the dependency from parts. For example, for parts = ["DateTime"], return "DateTime"
// Take care of fully qualified names and aliases
func (v *PhpVisitor) nameDependencyFromParts(parts []ast.Vertex) *pb.StmtExternalDependency {

	className := ""
	for _, part := range parts {
		className += string(part.(*ast.NamePart).Value) + "\\"
	}
	className = strings.TrimSuffix(className, "\\")
	dependency := &pb.StmtExternalDependency{
		ClassName: className,
	}

	// if is reserved keyword, return nil
	if v.IsReservedKeyword(dependency.ClassName) {
		return nil
	}

	dependency = v.unalias(dependency)
	dependency.ClassName = strings.TrimPrefix(dependency.ClassName, "\\")

	return dependency
}

// When a use statement is found, store the alias
func (v *PhpVisitor) StmtUse(node *ast.StmtUseList) {

	if v.aliases == nil {
		v.aliases = make(map[string]string)
	}

	for _, use := range node.Uses {

		// Get the alias
		alias := ""
		if use.(*ast.StmtUse).Alias != nil {
			alias = string(use.(*ast.StmtUse).Alias.(*ast.Identifier).Value)
		}

		// Get the full name (with namespace)
		parts := use.(*ast.StmtUse).Use.(*ast.Name).Parts
		name := ""
		for _, part := range parts {
			name += "\\" + string(part.(*ast.NamePart).Value)
		}

		// If alias is empty, use the short name as alias
		if alias == "" {
			// Get the last part of the name
			lastPart := parts[len(parts)-1].(*ast.NamePart).Value
			alias = string(lastPart)
		}

		// trim leading slash
		name = strings.TrimPrefix(name, "\\")
		alias = strings.TrimPrefix(alias, "\\")

		// Add to the list of aliases
		v.aliases[alias] = name
	}
}

// When a new object is created, store the dependency
func (v *PhpVisitor) ExprNew(node *ast.ExprNew) {

	if v.currentMethod == nil {
		return
	}
	if node.Class == nil {
		return
	}

	// Detect is node has ast.Name or ast.NameFullyQualified, to manage the nil pointer
	var parts []ast.Vertex
	switch node.Class.(type) {
	case *ast.Name:
		// usage of name, like new DateTime()
		parts = node.Class.(*ast.Name).Parts
	case *ast.NameFullyQualified:
		// usage of fully qualified name, like new \DateTime()
		parts = node.Class.(*ast.NameFullyQualified).Parts
		parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
	default:
		// Handle unexpected types
	}

	dependency := v.nameDependencyFromParts(parts)
	if dependency == nil {
		return
	}

	// Add dependency to method
	if v.currentMethod != nil {
		v.currentMethod.Stmts.StmtExternalDependencies = append(v.currentMethod.Stmts.StmtExternalDependencies, dependency)
	}

	// Add dependency to class
	if v.currentClass != nil {
		v.currentClass.Stmts.StmtExternalDependencies = append(v.currentClass.Stmts.StmtExternalDependencies, dependency)
	}
}

// When a method is called, store the dependency
func (v *PhpVisitor) ExprStaticCall(node *ast.ExprStaticCall) {

	calledFunctionName := ""
	switch node.Call.(type) {
	case *ast.Identifier:
		calledFunctionName = string(node.Call.(*ast.Identifier).Value)
	case *ast.ExprVariable:
		calledFunctionName = string(node.Call.(*ast.ExprVariable).Name.(*ast.Identifier).Value)
	default:
		// Handle unexpected types
	}

	// Detect is node has ast.Name or ast.NameFullyQualified, to manage the nil pointer
	var parts []ast.Vertex
	switch node.Class.(type) {
	case *ast.Name:
		// usage of name, like new DateTime()
		parts = node.Class.(*ast.Name).Parts
	case *ast.NameFullyQualified:
		// usage of fully qualified name, like new \DateTime()
		parts = node.Class.(*ast.NameFullyQualified).Parts
		parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
	default:
		// Handle unexpected types
	}

	dependency := v.nameDependencyFromParts(parts)
	if dependency == nil {
		return
	}
	dependency.FunctionName = calledFunctionName

	// Add dependency to method
	if v.currentMethod != nil {
		v.currentMethod.Stmts.StmtExternalDependencies = append(v.currentMethod.Stmts.StmtExternalDependencies, dependency)
	}

	// Add dependency to class
	if v.currentClass != nil {
		v.currentClass.Stmts.StmtExternalDependencies = append(v.currentClass.Stmts.StmtExternalDependencies, dependency)
	}
}

// When a method is called, store the dependency
func (v *PhpVisitor) ExprStaticPropertyFetch(node *ast.ExprStaticPropertyFetch) {

	// Detect is node has ast.Name or ast.NameFullyQualified, to manage the nil pointer
	var parts []ast.Vertex
	switch node.Class.(type) {
	case *ast.Name:
		// usage of name, like new DateTime()
		parts = node.Class.(*ast.Name).Parts
	case *ast.NameFullyQualified:
		// usage of fully qualified name, like new \DateTime()
		parts = node.Class.(*ast.NameFullyQualified).Parts
		parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
	default:
		// Handle unexpected types
	}

	dependency := v.nameDependencyFromParts(parts)
	if dependency == nil {
		return
	}

	// Add dependency to method
	if v.currentMethod != nil {
		v.currentMethod.Stmts.StmtExternalDependencies = append(v.currentMethod.Stmts.StmtExternalDependencies, dependency)
	}

	// Add dependency to class
	if v.currentClass != nil {
		v.currentClass.Stmts.StmtExternalDependencies = append(v.currentClass.Stmts.StmtExternalDependencies, dependency)
	}
}

// When class constant is called, store the dependency
func (v *PhpVisitor) ExprClassConstFetch(node *ast.ExprClassConstFetch) {

	// Detect is node has ast.Name or ast.NameFullyQualified, to manage the nil pointer
	var parts []ast.Vertex
	switch node.Class.(type) {
	case *ast.Name:
		// usage of name, like new DateTime()
		parts = node.Class.(*ast.Name).Parts
	case *ast.NameFullyQualified:
		// usage of fully qualified name, like new \DateTime()
		parts = node.Class.(*ast.NameFullyQualified).Parts
		parts = append([]ast.Vertex{&ast.NamePart{Value: []byte("\\")}}, parts...)
	default:
		// Handle unexpected types
	}

	dependency := v.nameDependencyFromParts(parts)
	if dependency == nil {
		return
	}

	// Add dependency to method
	if v.currentMethod != nil {
		v.currentMethod.Stmts.StmtExternalDependencies = append(v.currentMethod.Stmts.StmtExternalDependencies, dependency)
	}

	// Add dependency to class
	if v.currentClass != nil {
		v.currentClass.Stmts.StmtExternalDependencies = append(v.currentClass.Stmts.StmtExternalDependencies, dependency)
	}
}

// When property is typed, store the dependency
func (v *PhpVisitor) StmtProperty(node *ast.StmtProperty) {

	if v.currentClass == nil {
		return
	}

	prop := node.Var
	if prop == nil {
		return
	}

	// Ensure identifier
	_, ok := prop.(*ast.ExprVariable)
	if !ok {
		return
	}

	// Type cannot be reached with the current parser.
	// We need to get previous chars and stop to first blanc char
	startPos := prop.GetPosition().StartPos
	codes := v.linesOfFile
	// convert array of string to one string
	raw := strings.Join(codes, "\n")
	endPos := startPos
	isFirstSeparator := true
	lenRaw := len(raw)
	for i := startPos - 1; i >= 0; i-- {

		if i >= lenRaw {
			// Parser may break when the file contains specific regex, like:
			// if (preg_match('/^('.$pattern.'([ ]++|$))(.*+)/', $Line['text'], $matches))
			// We should open an issue to the parser, or try to fix it :(
			if v.file.Errors == nil {
				v.file.Errors = make([]string, 0)
			}
			v.file.Errors = append(v.file.Errors, "Parser error: i >= lenRaw in StmtProperty")
			return
		}

		if raw[i] == ' ' || raw[i] == '\t' {
			if !isFirstSeparator {
				break
			}
		} else {
			// when char is found
			isFirstSeparator = false
		}
		endPos = i
	}

	classname := raw[endPos:startPos]
	classname = strings.TrimSpace(classname)
	classname = strings.Trim(classname, "?")
	if classname == "" {
		return
	}

	// if classname is a reserved word (like int, string, etc), do not add it as dependency
	if v.IsReservedKeyword(classname) {
		return
	}

	dependency := &pb.StmtExternalDependency{
		ClassName: classname,
	}
	dependency = v.unalias(dependency)

	// Add dependency to class
	v.currentClass.Stmts.StmtExternalDependencies = append(v.currentClass.Stmts.StmtExternalDependencies, dependency)
}

func (v *PhpVisitor) IsReservedKeyword(expression string) bool {

	reserved := []string{"public", "private", "protected", "var", "int", "string", "float", "bool", "array", "object", "callable", "iterable", "void", "static", "self", "parent"}
	for _, r := range reserved {
		if r == expression {
			return true
		}
	}

	return false
}
