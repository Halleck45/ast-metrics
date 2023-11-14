<?php

namespace App;

use PHPUnit\Framework\TestCase;

class DumperTest extends TestCase
{

    public function tearDown(): void
    {
        if (file_exists(sys_get_temp_dir() . '/test.php')) {
            unlink(sys_get_temp_dir() . '/test.php');
        }
    }

    public function testDecisionPointsAreDumped()
    {
        $code = <<<EOT
<?php
if (\$a == \$b) {
    if (\$a1 == \$b1) {
        fiddle();
    } elseif (\$a2 == \$b2) {
        fiddle();
    } else {
        fiddle();
    }
} elseif (\$c == \$d) {
    while (\$c == \$d) {
        fiddle();
    }
} elseif (\$e == \$f) {
    for (\$n = 0; \$n < \$h; \$n++) {
        fiddle();
    }
} else {
    switch (\$z) {
        case 1:
            fiddle();
            break;
        case 2:
            fiddle();
            break;
        case 3:
            fiddle();
            break;
        default:
            fiddle();
            break;
    }
}
EOT;

        $expected = <<<EOT
{
    "stmtDecisionIf": [
        {
            "stmts": {
                "stmtDecisionIf": [
                    {
                        "stmts": {}
                    }
                ],
                "stmtDecisionElseIf": [
                    {
                        "stmts": {}
                    }
                ],
                "stmtDecisionElse": [
                    {
                        "stmts": {}
                    }
                ]
            }
        }
    ],
    "stmtDecisionElseIf": [
        {
            "stmts": {
                "stmtLoop": [
                    {
                        "stmts": {}
                    }
                ]
            }
        },
        {
            "stmts": {
                "stmtLoop": [
                    {
                        "stmts": {}
                    }
                ]
            }
        }
    ],
    "stmtDecisionElse": [
        {
            "stmts": {
                "stmtDecisionCase": [
                    {
                        "stmts": {}
                    },
                    {
                        "stmts": {}
                    },
                    {
                        "stmts": {}
                    },
                    {
                        "stmts": {}
                    }
                ],
                "stmtDecisionSwitch": [
                    {}
                ]
            }
        }
    ]
}
EOT;
        $expected = json_decode($expected, true);

        $filename = sys_get_temp_dir() . '/test.php';
        file_put_contents($filename, $code);

        // parse the code
        $parser = new FileParser();
        // with disable details in order to have a readable output
        $parser->enableDetails(false);
        $nodes = $parser->parse($filename);

        // dump the code
        $dumper = new Dumper($filename);
        $protoFile = $dumper->dump($nodes);

        // convert it to JSON
        $json = json_decode($protoFile->serializeToJsonString(), true);

        // compare the output
        $this->assertEquals($expected, $json['stmts']);
    }

    /**
     * @group loc
     */
    public function testLinesOfCodeAreCorrect()
    {
        $code = <<<EOT
<?php


/**
 * This is a comment
 */
class Foo {

    // bar
    private \$bar;
    
    /**
     * This is a comment
     */
    public function __construct()
    {
        // inline comment
        
        // inline comment
        \$this->bar = 1;
    }
}


/**
 * This is a comment
 */
class AnotherClass {
    public function __construct()
    {
        // inline comment
        \$this->bar = 1;
    }
}
EOT;

        $filename = sys_get_temp_dir() . '/test.php';
        file_put_contents($filename, $code);

        // parse the code
        $parser = new FileParser();
        // with disable details in order to have a readable output
        $nodes = $parser->parse($filename);

        // dump the code
        $dumper = new Dumper($filename);
        $protoFile = $dumper->dump($nodes);

        // convert it to JSON
        $json = json_decode($protoFile->serializeToJsonString(), true);

        // 20 lines of code
        // 9 lines of comments
        // 11 ncloc
        // 2 lines of logical code
        // Assertions on class
        $linesOfCode = $json['stmts']['stmtClass'][0]['linesOfCode'];
        $this->assertEquals(9, $linesOfCode['commentLinesOfCode']);
        $this->assertEquals(19, $linesOfCode['linesOfCode']);

        // Assertions on class
        $linesOfCode = $json['stmts']['stmtClass'][1]['linesOfCode'];
        $this->assertEquals(4, $linesOfCode['commentLinesOfCode']);
        $this->assertEquals(10, $linesOfCode['linesOfCode']);


        // Assertions on file
        $linesOfCode = $json['linesOfCode'];
        $this->assertEquals(13, $linesOfCode['commentLinesOfCode']);
        $this->assertEquals(29, $linesOfCode['linesOfCode']);
    }

    /**
     * @group operator
     */
    public function testOperatorsAreDumped()
    {
        $code = <<<EOT
<?php

class Foo {

    public function __construct()
    {
        \$a = 1;
        \$b = 2;
        \$c = \$a + \$b;
        \$d = \$a - \$b;
        \$e = \$a * \$b;
    }
}
EOT;

        $code = <<<EOT
<?php

class Foo {

    public function __construct()
    {
        \$a = 1;
        \$this->b = 2;
        \$c = \$a + \$b;
        \$d = \$a - \$this->b;
        \$e = \$a * \$b;
        \$e = \$a + \$b;
        
        \$c = \$a > \$b;
        \$this->foo->bar();
        \$this->foo->get('abc')->add(\$this);
    }
}
EOT;

        $filename = sys_get_temp_dir() . '/test.php';
        file_put_contents($filename, $code);

        // parse the code
        $parser = new FileParser();
        // with disable details in order to have a readable output
        $nodes = $parser->parse($filename);

        // dump the code
        $dumper = new Dumper($filename);
        $protoFile = $dumper->dump($nodes);

        // convert it to JSON
        $json = json_decode($protoFile->serializeToJsonString(), true);

        $method = $json['stmts']['stmtClass'][0]['stmts']['stmtFunction'][0];
        $operands = $method['operands'];
        $operators = $method['operators'];

        // operands as array of string
        $operands = array_map(function ($operand) {
            return $operand['name'];
        }, $operands);
        $this->assertEquals(['a', 'this->b', 'c', 'd', 'e', 'e', 'c'], $operands);

        // operators as array of string
        $operators = array_map(function ($operator) {
            return $operator['name'];
        }, $operators);
        $expected = [
            'Assign',
            'Assign',
            'Assign',
            'BinaryOp_Plus',
            'Assign',
            'BinaryOp_Minus',
            'Assign',
            'BinaryOp_Mul',
            'Assign',
            'BinaryOp_Plus',
            'Assign',
            'BinaryOp_Greater',
        ];
        $this->assertEquals($expected, $operators);
    }

}