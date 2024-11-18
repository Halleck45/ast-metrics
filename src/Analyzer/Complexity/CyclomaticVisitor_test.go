package Analyzer

import (
	"testing"

	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/Engine/Golang"
	"github.com/halleck45/ast-metrics/src/Engine/Php"
	"github.com/stretchr/testify/assert"
)

func TestItCalculateCyclomaticComplexityForGoLang(t *testing.T) {

	fileContent := `
    package main

    import "fmt"

    func example() {
        if true {
            if true {
                fmt.Println("Hello")
            }
        } else if true {
            fmt.Println("Hello")
        } else {
            fmt.Println("Hello")
        }
    }
    `

	parser := &Golang.GolangRunner{}
	pbFile, err := Engine.CreateTestFileWithCode(parser, fileContent)
	assert.Nil(t, err)

	visitor := CyclomaticComplexityVisitor{}
	ccn := visitor.Calculate(pbFile.Stmts)
	assert.Equal(t, int32(4), ccn)
}

func TestItCalculateCyclomaticComplexityForPhp(t *testing.T) {

	visitor := CyclomaticComplexityVisitor{}

	fileContent := `
    <?php
    namespace App;
    
    function example() {
        if (true) {
            if (true) {
                echo "Hello";
            }
        } else if (true) {
            echo "Hello";
        } else {
            echo "Hello";
        }
    }
    `

	parser := &Php.PhpRunner{}
	pbFile, err := Engine.CreateTestFileWithCode(parser, fileContent)
	assert.Nil(t, err)

	ccn := visitor.Calculate(pbFile.Stmts)
	assert.Equal(t, int32(4), ccn)
}

func TestItCalculateCyclomaticComplexityForComplexPhp(t *testing.T) {

	fileContent := `
    <?php
    namespace App;

    class Foo {
        
        public function example() {
            if (true) {
                if (true) {
                    echo "Hello";
                }
            } else if (true) {
                echo "Hello";
            } else {
                echo "Hello";
            }
        }

        public function example2() {
            if (true) {
                echo 'ok';
            } else {
                echo 'ko';
            }
        }
    }
    `

	parser := &Php.PhpRunner{}
	pbFile, err := Engine.CreateTestFileWithCode(parser, fileContent)
	assert.Nil(t, err)

	visitor := CyclomaticComplexityVisitor{}
	ccn := visitor.Calculate(pbFile.Stmts)
	assert.Equal(t, int32(6), ccn)
}

func TestItCalculateCyclomaticComplexityForMethodItself(t *testing.T) {

	fileContent := `
    <?php

    function example() {
        $a = 123;
        return $a;
    }
    `

	parser := &Php.PhpRunner{}
	pbFile, err := Engine.CreateTestFileWithCode(parser, fileContent)
	assert.Nil(t, err)

	visitor := CyclomaticComplexityVisitor{}
	ccn := visitor.Calculate(pbFile.Stmts)
	assert.Equal(t, int32(1), ccn)
}

func TestItCalculateCyclomaticComplexityForAllDecisionPoints(t *testing.T) {

	fileContent := `
    <?php

    namespace Foo;

    class Foo
    {
        public function example()
        {
            $a = 123;
            if ($a > 1) {
                if ($a > 2) {
                    // ...
                } elseif ($a > 3) {
                    // ...
                } else {
                    // ...
                }
            } elseif ($a > 4) {
                while ($a > 5) {
                    // ...
                }
                do {
                    // ...
                } while ($a > 6);
            } else {
                switch ($a) {
                    case 1:
                        // ...
                        break;
                    case 2:
                        // ...
                        break;
                    default:
                        // ...
                }
            }

            foreach ($a as $b) {
                // ...
            }

            for ($i = 0; $i < 10; $i++) {
                // ...
            }
        }
    }
    `

	parser := &Php.PhpRunner{}
	pbFile, err := Engine.CreateTestFileWithCode(parser, fileContent)
	assert.Nil(t, err)

	visitor := CyclomaticComplexityVisitor{}
	ccn := visitor.Calculate(pbFile.Stmts)
	assert.Equal(t, int32(13), ccn)
}
