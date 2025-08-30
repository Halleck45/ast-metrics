package Php

import (
	"fmt"
	"testing"

	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/stretchr/testify/assert"
)

func TestPhpRunner(t *testing.T) {
	phpSource := `
<?php
namespace Foo\Bar;

class calculatrice {
	// A single line comment is here
	// A single line comment is here

	public function add($a, $b) {
		// A single line comment is here
		// A single line comment is here
		// A single line comment is here
		// A single line comment is here
		return $a + $b;
	}


	/**
	 * Divide a by b
	 */
	public function divide(int $a, int $b) {
		if ($b == 0) {
			throw new \InvalidArgumentException('Division by zero.');
		}



		$d = $a / $b;
		$d += 1;
		$e = $this->add($this->a1, $d);
		return $e;
	}
}
`

	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure no error
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure functions
	assert.Equal(t, 0, len(result.Stmts.StmtFunction), "Incorrect number of functions")

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "calculatrice", class1.Name.Short, "Expected class name to be 'calculatrice', got %s", class1.Name)

	// Ensure functions
	assert.Equal(t, 2, len(class1.Stmts.StmtFunction), "Incorrect number of functions in class")

	func1 := class1.Stmts.StmtFunction[0]
	assert.Equal(t, "add", func1.Name.Short, "Expected function name to be 'add', got %s", func1.Name)
	assert.Equal(t, "Foo\\Bar\\calculatrice::add", func1.Name.Qualified, "Expected function name")
	func2 := class1.Stmts.StmtFunction[1]
	assert.Equal(t, "divide", func2.Name.Short, "Expected function name to be 'divide', got %s", func2.Name)
	assert.Equal(t, "Foo\\Bar\\calculatrice::divide", func2.Name.Qualified, "Expected function name")

	// Ensure operands
	// [name:"a" name:"b" name:"a" name:"b"]
	// Convert to string (for easier comparison)
	operandsAsString := fmt.Sprintf("%v", func1.Operands)
	operandsExpectedAsString := "[name:\"$a\" name:\"$b\" name:\"$a\" name:\"$b\"]"
	assert.Equal(t, operandsExpectedAsString, operandsAsString, "Expected operands to be %s, got %s", operandsExpectedAsString, operandsAsString)

	// Ensure operands of function 2
	// [a, b, b, d, a, b, d, e, a, d, e]
	// Convert to string (for easier comparison)
	operandsAsString = fmt.Sprintf("%v", func2.Operands)
	operandsExpectedAsString = "[name:\"$a\" name:\"$b\" name:\"$b\" name:\"$d\" name:\"$a\" name:\"$b\" name:\"$d\" name:\"$e\" name:\"$this->a1\" name:\"$d\" name:\"$e\"]"
	assert.Equal(t, operandsExpectedAsString, operandsAsString, "Expected operands to be %s, got %s", operandsExpectedAsString, operandsAsString)

	// Ensure operators
	// [+]
	// Convert to string (for easier comparison)
	operatorsAsString := fmt.Sprintf("%v", func1.Operators)
	operatorsExpectedAsString := "[name:\"+\"]"
	assert.Equal(t, operatorsExpectedAsString, operatorsAsString, "Expected operators to be %s, got %s", operatorsExpectedAsString, operatorsAsString)

	// Ensure operators of function 2
	// [==, / ]
	// Convert to string (for easier comparison)
	operatorsAsString = fmt.Sprintf("%v", func2.Operators)
	operatorsExpectedAsString = "[name:\"==\" name:\"/\" name:\"+=\"]"
	assert.Equal(t, operatorsExpectedAsString, operatorsAsString, "Expected operators to be %s, got %s", operatorsExpectedAsString, operatorsAsString)

	// Ensure LOC
	assert.Equal(t, int32(7), func1.LinesOfCode.LinesOfCode, "Expected LOC")
	assert.Equal(t, int32(1), func1.LinesOfCode.LogicalLinesOfCode, "Expected LLOC")
	assert.Equal(t, int32(4), func1.LinesOfCode.CommentLinesOfCode, "Expected CLOC")
	// Ensure LOC
	assert.Equal(t, int32(12), func2.LinesOfCode.LinesOfCode, "Expected LOC")
	assert.Equal(t, int32(7), func2.LinesOfCode.LogicalLinesOfCode, "Expected LLOC")
	assert.Equal(t, int32(3), func2.LinesOfCode.CommentLinesOfCode, "Expected CLOC")
}

func TestPhpLoops(t *testing.T) {
	phpSource := `
<?php

function test() {
	for ($i = 0; $i < 10; $i++) {
		echo $i;
	}

	foreach ($array as $value) {
		echo $value;
	}

	while ($i < 10) {
		echo $i;
		$i++;
	}

	do {
		echo $i;
		$i++;
	} while ($i < 10);
	}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure no error
	assert.Nil(t, err, "Expected no error, got %s", err)

	// 1 function should be found
	assert.Equal(t, 1, len(result.Stmts.StmtFunction), "Incorrect number of functions")

	// 4 loops should be found
	func1 := result.Stmts.StmtFunction[0]
	assert.Equal(t, 4, len(func1.Stmts.StmtLoop), "Incorrect number of loops")
}

func TestEnumWithoutNamespace(t *testing.T) {
	phpSource := `
<?php

enum Values {
	case A;
	case B;
	case C;

	public function __toString() {
		return match($this) {
			Values::A => 'A',
			Values::B => 'B',
			Values::C => 'C',
		};
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	// Ensure no error
	assert.Nil(t, err, "Expected no error, got %s", err)

	// a class (enum) should be found
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "Values", class1.Name.Short, "Expected class name to be 'Values', got %s", class1.Name)
	assert.Equal(t, "Values", class1.Name.Qualified, "Expected class name to be 'Values', got %s", class1.Name.Qualified)

	// one method should be found
	assert.Equal(t, 1, len(class1.Stmts.StmtFunction), "Incorrect number of functions in class")
	func1 := class1.Stmts.StmtFunction[0]
	assert.Equal(t, "__toString", func1.Name.Short, "Expected function name to be '__toString', got %s", func1.Name)
}

func TestTrait(t *testing.T) {
	phpSource := `
<?php

trait MonTrait1 {
	public function foo() {
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure no error
	assert.Nil(t, err, "Expected no error, got %s", err)

	// a class (trait) should be found
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "MonTrait1", class1.Name.Short, "Expected class name to be 'MonTrait1', got %s", class1.Name)

	// one method should be found
	assert.Equal(t, 1, len(class1.Stmts.StmtFunction), "Incorrect number of functions in class")
	func1 := class1.Stmts.StmtFunction[0]
	assert.Equal(t, "foo", func1.Name.Short, "Expected function name to be 'foo', got %s", func1.Name)
}

func TestPhpInterface(t *testing.T) {
	phpSource := `
<?php

namespace Truc;

interface Foo {
	public function bar();
}
`

	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Check that a namespace is found
	assert.Equal(t, 1, len(result.Stmts.StmtNamespace), "Incorrect number of namespaces")

	// one interface should be found
	assert.Equal(t, 1, len(result.Stmts.StmtInterface), "Incorrect number of interfaces")
	class1 := result.Stmts.StmtInterface[0]
	assert.Equal(t, "Foo", class1.Name.Short, "Expected class name to be 'Foo', got %s", class1.Name)
	assert.Equal(t, "Truc\\Foo", class1.Name.Qualified, "Expected class name to be 'Truc\\Foo', got %s", class1.Name.Qualified)
}

func TestPhpOperators(t *testing.T) {
	phpSource := `
<?php

function test() {
	$a = 1 + 2;
	$b = 1 - 2;
	$c = 1 * 2;
	$d = 1 / 2;
	$e = 1 % 2;
	$f = 1 ** 2;
	$g = 1 . 2;
	$h = 1 << 2;
	$i = 1 >> 2;
	$j = 1 & 2;
	$k = 1 | 2;
	$l = 1 ^ 2;
	$m = 1 && 2;
	$n = 1 || 2;
	$o = 1 ?? 2;
	$p = 1 == 2;
	$q = 1 === 2;
	$r = 1 != 2;
	$s = 1 !== 2;
	$t = 1 < 2;
	$u = 1 <= 2;
	$v = 1 > 2;
	$w = 1 >= 2;
	$x = 1 <=> 2;
	// bitwise operators
	$ab = $a &= $b;
	$ab = $a |= $b;
	$ab = $a ^= $b;
	$ab = $a ??= $b;
	$ab = $a .= $b;
	$ab = $a /= $b;
	$ab = $a -= $b;
	$ab = $a %= $b;
	$ab = $a *= $b;
	$ab = $a += $b;
	$ab = $a **= $b;
	$ab = $a <<= $b;
	$ab = $a >>= $b;
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// 1 function should be found
	assert.Equal(t, 1, len(result.Stmts.StmtFunction), "Incorrect number of functions")

	// operators should be found
	func1 := result.Stmts.StmtFunction[0]
	assert.Equal(t, 36, len(func1.Operators), "Incorrect number of operators")
}

func TestPhpIfCases(t *testing.T) {
	phpSource := `
<?php
function foo() {
	if ($a == 1) {
		echo "a";
	} elseif ($a == 2) {
		echo "b";
	} else {
		echo "c";
	}

	if ($a == 1) {
		echo "a";
	} else {
		echo "b";
	}

	if ($a == 1) {
		echo "a";
	}
}

function bar() {
	if ($a == 1) {
		if ($b == 2) {
			echo "a";
		} elseif ($b == 3) {
			echo "b";
		}
	} elseif ($a == 2) {
		echo "c";
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure functions
	assert.Equal(t, 2, len(result.Stmts.StmtFunction), "Incorrect number of functions")

	// Function 1
	func1 := result.Stmts.StmtFunction[0]
	assert.Equal(t, 4, len(func1.Stmts.StmtDecisionIf), "Incorrect number of if statements")

	// Function 2
	func2 := result.Stmts.StmtFunction[1]
	assert.Equal(t, 4, len(func2.Stmts.StmtDecisionIf), "Incorrect number of if statements")
}

func TestNamesapceWithoutName(t *testing.T) {
	phpSource := `
<?php

namespace {
    class Foo
    {
        public function __construct()
        {
            echo 'Foo::__construct()';
        }
    }
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "Foo", class1.Name.Short, "Expected class name to be 'Foo', got %s", class1.Name)

}

func TestNonValidFile(t *testing.T) {

	phpSource := `
<?php

class Foo 
{
{
	public function foo()
	{
	}
}
`

	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure errors
	assert.NotEmpty(t, result.Errors)
}

func TestNonUtf8Classnames(t *testing.T) {

	// create a non-utf8 classname
	classname := []byte{0x80, 0x80, 0x80, 0x80, 0x80}

	phpSource := `
<?php

class ` + string(classname) + `
{
	public function foo()
	{
	}
}
`

	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure errors
	assert.Empty(t, result.Errors)

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "@non-utf8", class1.Name.Short, "Expected class name to be '@non-utf8', got %s", class1.Name)
}

func TestExternalCallsInNamespace(t *testing.T) {

	phpSource := `
<?php

namespace My\Namespace;

use External\Class1;
use External\Class2 as ClassAliased;
use External\Class3 as StaticClassAliased;
use stdClass;

class TestedClass
{
	private ?\FullNamespace\Class1   $a1; // Type hinted
	private string   $a2;
	private ?string   $a2;
	private $b1;
	var $c1;

	public function foo1(?LocalClass2 $a) : ?LocalClass3
	{
		// Type hinted + New instance
		$o = new LocalClass5;
		$o = new ClassAliased();
		throw new \InvalidArgumentException('Division by zero.');
		return $o->bar();
	}

	public function foo2(stdClass $a, $b) : LocalClass1
	{
		// return type hinted + Static call
		LocalClass4::externalMethod();
		StaticClassAliased::anotherMethod();
		$x = LocalClass5::$ATTRIBUTE1;
		$y = \Fully\Qualified1\Class1::CONSTANT1;
		$ya = \Fully\Qualified1\Class2::$ATTRIBUTE1;
		$z = \Fully\Qualified1\Class3::method1();

		$this->foo1();
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)
	assert.Empty(t, result.Errors)

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "TestedClass", class1.Name.Short, "Expected class name to be 'TestedClass', got %s", class1.Name)

	// Ensure external calls
	dependencies := class1.Stmts.StmtExternalDependencies
	expected := []string{
		"FullNamespace\\Class1",
		"My\\Namespace\\LocalClass2",
		"My\\Namespace\\LocalClass3",
		"External\\Class2",
		"My\\Namespace\\LocalClass5",
		"External\\Class3",
		"InvalidArgumentException",
		"stdClass",
		"My\\Namespace\\LocalClass1",
		"My\\Namespace\\LocalClass4",
		"My\\Namespace\\LocalClass5",
		"Fully\\Qualified1\\Class1",
		"Fully\\Qualified1\\Class2",
		"Fully\\Qualified1\\Class3",
	}

	found := []string{}
	for _, dep := range dependencies {
		found = append(found, dep.ClassName)
	}

	// Compare the list
	assert.ElementsMatch(t, expected, found, "Incorrect external dependencies")
}

func TestExternalCallsInRootNamespace(t *testing.T) {

	phpSource := `
<?php

use External\Class1;
use External\Class2 as ClassAliased;
use External\Class3 as StaticClassAliased;
use stdClass;

class TestedClass
{
	private \FullNamespace\Class1   $a1; // Type hinted
	private string   $a2;
	private $b1;
	var $c1;

	public function foo1(LocalClass2 $a)
	{
		// Type hinted + New instance
		$o = new LocalClass5;
		$o = new ClassAliased();
		throw new \InvalidArgumentException('Division by zero.');
		return $o->bar();
	}

	public function foo2(stdClass $a, $b) : LocalClass1
	{
		// return type hinted + Static call
		LocalClass4::externalMethod();
		StaticClassAliased::anotherMethod();
		$x = LocalClass5::$ATTRIBUTE1;
		$y = \Fully\Qualified1\Class1::CONSTANT1;
		$ya = \Fully\Qualified1\Class2::$ATTRIBUTE1;
		$z = \Fully\Qualified1\Class3::method1();

		$this->foo1();
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)
	assert.Empty(t, result.Errors)

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "TestedClass", class1.Name.Short, "Expected class name to be 'TestedClass', got %s", class1.Name)

	// Ensure external calls
	dependencies := class1.Stmts.StmtExternalDependencies
	expected := []string{
		"FullNamespace\\Class1",
		"LocalClass2",
		"External\\Class2",
		"LocalClass5",
		"External\\Class3",
		"InvalidArgumentException",
		"stdClass",
		"LocalClass1",
		"LocalClass4",
		"LocalClass5",
		"Fully\\Qualified1\\Class1",
		"Fully\\Qualified1\\Class2",
		"Fully\\Qualified1\\Class3",
	}

	found := []string{}
	for _, dep := range dependencies {
		found = append(found, dep.ClassName)
	}

	// Compare the list
	assert.ElementsMatch(t, expected, found, "Incorrect external dependencies")
}

func TestSwitchCasesAreCorrectlyParser(t *testing.T) {
	phpSource := `
<?php
function foo() {
	switch ($a) {
		case 1:
			echo "a";
			break;
		case 2:
			echo "b";
			break;
		default:
			echo "c";
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure functions
	assert.Equal(t, 1, len(result.Stmts.StmtFunction), "Incorrect number of functions")

	// Function 1
	func1 := result.Stmts.StmtFunction[0]
	assert.Equal(t, 1, len(func1.Stmts.StmtDecisionSwitch), "Incorrect number of switch statements")
}

func TestPhp84SyntaxPropertyHooks(t *testing.T) {
	phpSource := `<?php
class HasAuthors
{
    public string $credits { get; }
    public Author $mainAuthor { get; set; }
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "HasAuthors", class1.Name.Short, "Expected class name to be 'HasAuthors', got %s", class1.Name)
}

func TestPhp84NewIthoutParenthesis(t *testing.T) {
	phpSource := `<?php
class A {
	public function bar() {
		$name = new ReflectionClass($objectOrClass)->getShortName();
		$y = new Z;
		return $name;
	}
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure functions
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, 1, len(class1.Stmts.StmtFunction), "Incorrect number of functions")
	func1 := class1.Stmts.StmtFunction[0]
	assert.Equal(t, "bar", func1.Name.Short, "Expected function name to be 'bar', got %s", func1.Name)

	dependencies := class1.Stmts.StmtExternalDependencies
	expected := []string{
		"ReflectionClass",
		"Z",
	}

	found := []string{}
	for _, dep := range dependencies {
		found = append(found, dep.ClassName)
	}

	// Compare the list
	assert.ElementsMatch(t, expected, found, "Incorrect external dependencies")
}

func TestReadonlyClass(t *testing.T) {
	phpSource := `<?php
readonly class B {
}
`
	result, err := Engine.CreateTestFileWithCode(&PhpRunner{}, phpSource)
	assert.Nil(t, err, "Expected no error, got %s", err)

	// Ensure classes
	assert.Equal(t, 1, len(result.Stmts.StmtClass), "Incorrect number of classes")
	class1 := result.Stmts.StmtClass[0]
	assert.Equal(t, "B", class1.Name.Short, "Expected class name to be 'B', got %s", class1.Name)
}
