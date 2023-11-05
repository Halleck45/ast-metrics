<?php

namespace App;

use PHPUnit\Framework\TestCase;

class DumperTest extends TestCase {

    public function tearDown(): void
    {
        if(file_exists(sys_get_temp_dir() . '/test.php')) {
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
        $json = json_decode($protoFile->serializeToJsonString() , true);

        // compare the output
        $this->assertEquals($expected, $json['stmts']);
    }
}