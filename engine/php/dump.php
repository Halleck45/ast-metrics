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

$lexer = new PhpParser\Lexer\Emulative(['usedAttributes' => [
    'startLine', 'endLine', 'startFilePos', 'endFilePos', 'comments'
]]);
$parser = (new PhpParser\ParserFactory)->create(
    PhpParser\ParserFactory::PREFER_PHP7,
    $lexer
);
$dumper = new PhpParser\NodeDumper([
    'dumpComments' => true,
    'dumpPositions' => true,
]);
$prettyPrinter = new PhpParser\PrettyPrinter\Standard;

$traverser = new PhpParser\NodeTraverser();
$traverser->addVisitor(new PhpParser\NodeVisitor\NameResolver);


if (!file_exists($file)) {
    fwrite(STDERR, "File $file does not exist.\n");
    exit(1);
}

$code = file_get_contents($file);
$stmts = $parser->parse($code);


/**
 * According to the type of statement, we create the corresponding proto node and inject it into the parent node.
 *
 * @param \PhpParser\Node\Stmt $stmt
 * @param \NodeType\Stmt $parent
 * @return \NodeType\StmtClass|\NodeType\StmtFunction|\NodeType\StmtNamespace|null
 */
function stmtToProto(\PhpParser\Node\Stmt $stmt, \NodeType\Stmt $parent) {

    // Here is the list of supported statements
    // We factory the corresponding proto node for each of them
    switch(get_class($stmt)) {
        case \PhpParser\Node\Stmt\Namespace_::class:
            $protoNode = new \NodeType\StmtNamespace();
            $parent->setStmtNamespace($protoNode);
            break;
        case \PhpParser\Node\Stmt\Class_::class:
            $protoNode = new \NodeType\StmtClass();
            $parent->setStmtClass($protoNode);
            break;
        case \PhpParser\Node\Stmt\Function_::class:
        case \PhpParser\Node\Stmt\ClassMethod::class:
            $protoNode = new \NodeType\StmtFunction();
            $parent->setStmtFunction($protoNode);
            break;
        case \PhpParser\Node\Stmt\If_::class:
            $protoNode = new \NodeType\StmtDecisionIf();
            $parent->setStmtDecisionIf($protoNode);
            break;
        case \PhpParser\Node\Stmt\ElseIf_::class:
            $protoNode = new \NodeType\StmtDecisionElseIf();
            $parent->setStmtDecisionElseIf($protoNode);
            break;
        case \PhpParser\Node\Stmt\Else_::class:
            $protoNode = new \NodeType\StmtDecisionElse();
            $parent->setStmtDecisionElse($protoNode);
            break;
        case \PhpParser\Node\Stmt\Case_::class:
            $protoNode = new \NodeType\StmtDecisionCase();
            $parent->setStmtDecisionCase($protoNode);
            break;
        default:
            // not supported yet
            if(getenv('DEBUG')) {
                trigger_error("Not supported yet: " . get_class($stmt), E_USER_WARNING);
            }
            return null;
    }

    // Determine the name if the statement has one
    if(isset($stmt->name)) {
        $protoNode->setName(new \NodeType\Name([
            'short' => $stmt->name->toString(),
            'qualified' => $stmt->name->toString(),
        ]));
    }

    // Location (in code)
    $location = new \NodeType\StmtLocationInFile([
        'startLine' => $stmt->getStartLine(),
        'endLine' => $stmt->getEndLine(),
        'startFilePos' => $stmt->getAttribute('startFilePos'),
        'endFilePos' => $stmt->getAttribute('endFilePos'),
    ]);
    $protoNode->setLocation($location);

    // if contains sub statements, do the same for each of them
    if(property_exists($stmt, 'stmts')) {
        $stmts = (array) $stmt->stmts;
        $protoStmts = new \NodeType\Stmts();
        $protoNode->setStmts($protoStmts);
        $subs = [];
        foreach($stmts as $stmt) {
            $protoStmt = new \NodeType\Stmt();
            $added = stmtToProto($stmt, $protoStmt);
            if(!$added) {
                continue;
            }
            $subs[] = $protoStmt;
        }
        $protoStmts->setStmts($subs);
        $protoNode->setStmts($protoStmts);
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
foreach($stmts as $stmt) {

    $nodeStmt = new \NodeType\Stmt();

    // convert to proto
    $added = stmtToProto($stmt, $nodeStmt);
    if(!$added) {
        continue;
    }

    $subs[] = $nodeStmt;
}

$protoStmts->setStmts($subs);
$format = getenv('OUTPUT_FORMAT') ?: 'binary';
switch($format) {

    case 'json-pretty':
        echo json_encode(json_decode($fileNode->serializeToJsonString()), JSON_PRETTY_PRINT);
        break;
    case 'binary':
        echo $fileNode->serializeToString();
        break;
    case 'json':
    default:
        echo $fileNode->serializeToJsonString();
    break;
}

