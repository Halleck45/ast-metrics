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
$prettyPrinter = new PhpParser\PrettyPrinter\Standard;
$json = json_decode(json_encode($stmts), true);

// This main node describe the file itself
$fileNode = new \NodeType\File([
    'path' => realpath($file),
]);
$protoStmts = new \NodeType\Stmts();
$fileNode->setStmts($protoStmts);


function stmtFactory($stmt) {
    switch($stmt['nodeType']) {
        case 'Stmt_Namespace':
            $node = new \NodeType\StmtNamespace();
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
            return null;
    }

    // Determine the name if the statement has one
    if(isset($stmt['name'])){
        $name = $stmt['name'];
        $parts = $stmt['name']['parts'] ?? [];
        if (!empty($parts)) {
            $name = implode('', $stmt['name']['parts']);
        }
        if (!empty($stmt['name']['name'])) {
            $name = $stmt['name']['name'];
        }
        $node->setName(new \NodeType\Name([
            'short' => $name,
            'qualified' => $name,
        ]));
    }

    // Location (in code)
    $location = new \NodeType\StmtLocationInFile([
        'startLine' => $stmt['attributes']['startLine'],
        'endLine' => $stmt['attributes']['endLine'],
        'startFilePos' => $stmt['attributes']['startFilePos'],
        'endFilePos' => $stmt['attributes']['endFilePos'],
    ]);
    $node->setLocation($location);

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
file_put_contents('test.json', json_encode($json, JSON_PRETTY_PRINT));
foreach($json as $stmt) {
    stmtToProto($stmt, $protoStmts);
}

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

