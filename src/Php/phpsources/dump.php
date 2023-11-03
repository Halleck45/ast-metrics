#!/usr/bin/env php
<?php

// The code is directly inspired from nikic/php-parser/bin/php-parse
//
// The main difference is that we don't dump the JSON as it is, but we standardize it with our protobuf schemas.

require_once __DIR__ . '/vendor/autoload.php';
spl_autoload_register(function ($class) {
    $class = str_replace('\\', '/', $class);
    require_once __DIR__ . "/generated/$class.php";
});

ini_set('xdebug.max_nesting_level', 3000);
ini_set('xdebug.var_display_max_children', -1);
ini_set('xdebug.var_display_max_data', -1);
ini_set('xdebug.var_display_max_depth', -1);

if(!isset($argv[1])) {
    echo "Usage: DEBUG=1 OUTPUT_FORMAT=json php dump.php <file>\n";
    exit(1);
}
$file = (string) $argv[1];

if (!file_exists($file)) {
    fwrite(STDERR, "File $file does not exist.\n");
    exit(1);
}

$lexer = new PhpParser\Lexer\Emulative(['usedAttributes' => [
    'startLine', 'endLine', 'startFilePos', 'endFilePos', 'comments'
]]);
$parser = (new PhpParser\ParserFactory)->create(
    PhpParser\ParserFactory::PREFER_PHP7,
    $lexer
);
$traverser = new PhpParser\NodeTraverser();
$traverser->addVisitor(new PhpParser\NodeVisitor\NameResolver);

$code = file_get_contents($file);
$linesOfCode = explode(PHP_EOL, $code);
$stmts = $parser->parse($code);
$json = (array)json_decode(json_encode($stmts), true);

// This main node describe the file itself
$fileNode = new \NodeType\File([
    'path' => realpath($file),
]);
$protoStmts = new \NodeType\Stmts();
$fileNode->setStmts($protoStmts);
$aliases = [];
$comments = [];
function stmtFactory($stmt) {

    global $lastStructuredParentStmt; // current class, namespace, function, etc...
    global $linesOfCode; // current source
    global $aliases; // uses and aliases

    $node = null;
    switch($stmt['nodeType'] ?? null) {
        case 'Stmt_Namespace':
            $node = new \NodeType\StmtNamespace();
            $aliases = [];
            break;
        case 'Stmt_Class':
            $node = new \NodeType\StmtClass();
            break;
        case 'Stmt_Function':
        case 'Stmt_ClassMethod':
            $node = new \NodeType\StmtFunction();
            break;
        case 'Stmt_If':
            $node = new \NodeType\StmtDecisionIf();
            break;
        case 'Stmt_Use':
            foreach ($stmt['uses'] as $use) {
                $alias = nameType($use['alias'] ?? null);
                $name = nameType($use['name']);
                $aliases[$alias] = $name;
            }
            break;
        case 'Stmt_ElseIf':
            $node = new \NodeType\StmtDecisionElseIf();
            break;
        case 'Stmt_Else':
            $node = new \NodeType\StmtDecisionElse();
            break;
        case 'Stmt_Case':
            $node = new \NodeType\StmtDecisionCase();
            break;
        case 'Stmt_For':
        case 'Stmt_Foreach':
        case 'Stmt_While':
            $node = new \NodeType\StmtLoop();
            break;
        default:
            break;
    }

    // Expressions
    // Operators and operands
    // We don't need to store the details for all statements, so we only store the details for the parent struct
    //
    // Operators is a raw list of operators, like "=" or "++
    // Operands is a list of variables, like "$a" or "$b"
    // It's useful to calculate Halsdstead's complexity
    if ($lastStructuredParentStmt
        && (strpos($stmt['nodeType'] ?? '', 'Stmt_Expr') !== false || isset($stmt['expr']))
    ) {
        $operator= new \NodeType\StmtOperator([
                'name' => $stmt['expr']['nodeType'] ?? $stmt['nodeType'] ?? $stmt['expr'] ?? null,
        ]);
        $operators = $lastStructuredParentStmt->getOperators() ?? [];
        $operators[] = $operator;
        $lastStructuredParentStmt->setOperators($operators);
    }
    // Operands
    if ($lastStructuredParentStmt && ((isset($stmt['expr']['var']) || isset($stmt['var'])))) {
        $name = nameVar($stmt);
        if($name) {
            $operand = new \NodeType\StmtOperand([
                'name' => $name,
            ]);
            $operands = $lastStructuredParentStmt->getOperands() ?? [];
            $operands[] = $operand;
            $lastStructuredParentStmt->setOperands($operands);
        }

    }

    if($lastStructuredParentStmt) {
        // External uses (new, static, etc.)
        $exprs = array_merge($stmt['exprs'] ?? [], (array)($stmt['expr'] ?? []));
        foreach($exprs as $expr) {
            $usages = $lastStructuredParentStmt->getExternals() ?? [];
            if(isset($expr['class'])) {
                $name = nameType($expr['class']);
                $usages[] = new \NodeType\Name([
                    'short' => $name,
                    'qualified' => $name,
                ]);
            }
            $lastStructuredParentStmt->setExternals($usages);
        }
    }

    if (!$node) {
        return null;
    }

    // Determine the name if the statement has one
    if(isset($stmt['name'])){
        $name = nameType($stmt);
        $node->setName(new \NodeType\Name([
            'short' => $name,
            'qualified' => $name,
        ]));
    }

    // Extends and implements
    if(isset($stmt['extends'])){
        $name = nameType($stmt['extends']);
        $extends = $node->getExtends() ?? [];
        $extends[] = new \NodeType\Name([
            'short' => $name,
            'qualified' => $name,
        ]);
        $node->setExtends($extends);
    }

    // Implements
    if(isset($stmt['implements'])){
        $implements = $node->getImplements() ?? [];
        foreach($stmt['implements'] as $implement) {
            $name = nameType($implement);
            $implements[] = new \NodeType\Name([
                'short' => $name,
                'qualified' => $name,
            ]);
        }
        $node->setImplements($implements);
    }

    // Parameters (for functions and methods)
    if(isset($stmt['params'])){
        $parameters = $node->getParameters() ?? [];
        foreach($stmt['params'] as $param) {
            $type = nameType($param['type'] ?? null);
            $parameters[] = new \NodeType\StmtParameter([
                'name' => nameVar($param),
                'type' => new \NodeType\Name([
                    'short' => $type,
                    'qualified' => $type,
                ]),
            ]);
        }
        $node->setParameters($parameters);
    }


    // count blank lines in statement
    $concernedLines = array_slice($linesOfCode, $stmt['attributes']['startLine'] - 1, $stmt['attributes']['endLine'] - $stmt['attributes']['startLine'] + 1);
    // Location (in code)
    $location = new \NodeType\StmtLocationInFile([
        'startLine' => $stmt['attributes']['startLine'],
        'endLine' => $stmt['attributes']['endLine'],
        'startFilePos' => $stmt['attributes']['startFilePos'],
        'endFilePos' => $stmt['attributes']['endFilePos'],
        'blankLines' => count(array_filter($concernedLines, function($line) {
            return trim($line) === '';
        })),
    ]);
    $node->setLocation($location);

    // Determine if the statement is a decision or a structure
    if (method_exists($node, 'setComments')) {
        $lastStructuredParentStmt = $node;
    }


    if(!empty($stmt['attributes']['comments'])) {
        if ($lastStructuredParentStmt) {
            // Node is a class or a method
            $comments = $lastStructuredParentStmt->getComments() ?? [];
            $stmtsComments = $stmt['attributes']['comments'] ?? [];

            foreach ($stmtsComments as $comment) {
                $protoComment = new \NodeType\StmtComment([
                    //'text' => $comment['text'], // commented: today we don't need the text
                ]);
                $location = new \NodeType\StmtLocationInFile([
                    'startLine' => $comment['line'],
                    'endLine' => $comment['endLine'],
                    'startFilePos' => $comment['filePos'],
                    'endFilePos' => $comment['endFilePos'],
                ]);
                $protoComment->setLocation($location);
                $comments[] = $protoComment;
            }

            $lastStructuredParentStmt->setComments($comments);
        }
    }

    if(!empty($stmt['attributes']['comments'])) {
        if ($lastStructuredParentStmt) {
            // Node is a class or a method
            $comments = $lastStructuredParentStmt->getComments() ?? [];
            $stmtsComments = $stmt['attributes']['comments'] ?? [];

            foreach ($stmtsComments as $comment) {
                $protoComment = new \NodeType\StmtComment([
                    //'text' => $comment['text'], // commented: today we don't need the text
                ]);
                $location = new \NodeType\StmtLocationInFile([
                    'startLine' => $comment['line'],
                    'endLine' => $comment['endLine'],
                    'startFilePos' => $comment['filePos'],
                    'endFilePos' => $comment['endFilePos'],
                ]);
                $protoComment->setLocation($location);
                $comments[] = $protoComment;
            }

            $lastStructuredParentStmt->setComments($comments);
        }
    }

    return $node;
}


/**
 * According to the type of statement, we create the corresponding proto node and inject it into the parent node.
 *
 * @param \PhpParser\Node\Stmt $stmt
 * @param \NodeType\Stmt $parent
 * @return \NodeType\StmtClass|\NodeType\StmtFunction|\NodeType\StmtNamespace|null
 */
function stmtToProto(array $stmt, \NodeType\Stmts $parent) {
    $protoNode = stmtFactory($stmt);
    if(!$protoNode) {
        return null;
    }
    $collection = 'get' . str_replace('NodeType\\', '', get_class($protoNode));
    $parent->$collection()[] = $protoNode;


    // if contains sub statements, do the same for each of them
    $subStatements = array_filter(array_merge(
        $stmt['stmts'] ?? [],
        $stmt['cases'] ?? [],
        $stmt['else'] ?? [],
            $stmt['elseifs'] ?? [],
            [$stmt['stmt']?? []]
    ));
    if(!empty($subStatements) && is_array($subStatements)) {
        $protoStmts = new \NodeType\Stmts();
        $protoNode->setStmts($protoStmts);
        foreach($subStatements as $stmt) {
            if (!is_array($stmt)) {
                continue;
            }

            stmtToProto($stmt, $protoStmts);
        }
    }

    return $protoNode;
}
// This main node describe the file itself
$fileNode = new \NodeType\File([
    'path' => realpath($file),
]);
$protoStmts = new \NodeType\Stmts();
$fileNode->setStmts($protoStmts);
$subs = [];
if(getenv('DEBUG')) {
    file_put_contents('tmp.json', json_encode($json, JSON_PRETTY_PRINT));
}
foreach($json as $stmt) {
    stmtToProto($stmt, $protoStmts);
}

function nameVar($what) {
    if(isset($what['var'])) {
        return nameVar($what['var']);
    }
    if(isset($what['expr'])) {
        return nameVar($what['expr']);
    }

    return $what['name'] ?? null;
};

function nameType($what) {
    if(isset($what['type'])) {
        return nameType($what['type']);
    }

    $name = $what['name'] ?? null;
    $parts = $what['parts'] ?? [];
    if (!empty($parts)) {
        return unalias(implode('\\', $what['parts']));
    }

    if(is_array($name)) {
        $name = nameType($name);
    }

    return unalias($name);
};

function unalias($name) {
    global $aliases;
    if (isset($aliases[$name])) {
        return $aliases[$name];
    }

    return $name;
}

$format = getenv('OUTPUT_FORMAT') ?: 'binary';
switch($format) {

    case 'json-pretty':
        echo json_encode(json_decode($fileNode->serializeToJsonString()), JSON_PRETTY_PRINT);
        break;
    case 'binary':
        echo $fileNode->serializeToString();
        break;
    case 'raw':
        echo json_encode($json, JSON_PRETTY_PRINT);
        break;
    case 'null':
        break;
    case 'json':
    default:
        echo $fileNode->serializeToJsonString();
    break;
}