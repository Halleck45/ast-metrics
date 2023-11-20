#!/usr/bin/env php
<?php

// The code is directly inspired from nikic/php-parser/bin/php-parse
//
// The main difference is that we don't dump the JSON as it is, but we standardize it with our protobuf schemas.

namespace App;

use Google\Protobuf\Internal\RepeatedField;
use NodeType\File;
use NodeType\LinesOfCode;
use NodeType\StmtNamespace;


class Dumper
{
    private ?RepeatedField $comments;
    private array $aliases = [];
    private string $file;
    private $lastStructuredParentStmt;
    private string $lastNamespace = '';
    private array $linesOfCode = [];

    public function __construct(string $file)
    {
        $this->file = $file;
    }

    public function dump(array $json): File
    {

        // This main node describe the file itself
        $fileNode = new File([
            'path' => realpath($this->file),
            'programmingLanguage' => 'PHP'
        ]);
        $protoStmts = new \NodeType\Stmts();
        $fileNode->setStmts($protoStmts);
        $loc = new LinesOfCode();
        $fileNode->setLinesOfCode($loc);
        $this->lastStructuredParentStmt = $fileNode;
        $this->linesOfCode = file($this->file);

        if (getenv('DEBUG')) {
            file_put_contents('tmp.json', json_encode($json, JSON_PRETTY_PRINT));
        }

        foreach ($json as $stmt) {
            $addedNode = $this->stmtToProto($stmt, $protoStmts);

            if (!$addedNode) {
                continue;
            }

            if (method_exists($addedNode, 'getLinesOfCode') && $addedNode->getLinesOfCode() instanceof LinesOfCode) {
                $loc->setLinesOfCode($loc->getLinesOfCode() + $addedNode->getLinesOfCode()->getLinesOfCode());
                $loc->setLogicalLinesOfCode($loc->getLogicalLinesOfCode() + $addedNode->getLinesOfCode()->getLogicalLinesOfCode());
                $loc->setCommentLinesOfCode($loc->getCommentLinesOfCode() + $addedNode->getLinesOfCode()->getCommentLinesOfCode());
                $loc->setNonCommentLinesOfCode($loc->getNonCommentLinesOfCode() + $addedNode->getLinesOfCode()->getNonCommentLinesOfCode());
            }
        }

        if (getenv('DEBUG')) {
            file_put_contents('tmp.json', json_encode($json, JSON_PRETTY_PRINT));
        }

        return $fileNode;
    }


    private function stmtFactory(array $stmt)
    {
        $node = null;
        switch ($stmt['nodeType'] ?? null) {
            case 'Stmt_Namespace':
                $node = new \NodeType\StmtNamespace();
                $this->lastNamespace = $this->nameType($stmt['name']) ?? '\\';
                $this->aliases = [];
                break;
            case 'Stmt_Class':
                $node = new \NodeType\StmtClass();
                break;
            case 'Stmt_Function':
            case 'Stmt_ClassMethod':
                $node = new \NodeType\StmtFunction();
                $this->lastStructuredParentStmt = $node;
                break;
            case 'Stmt_If':
                $node = new \NodeType\StmtDecisionIf();
                break;
            case 'Stmt_Use':
                foreach ($stmt['uses'] as $use) {
                    $alias = $this->nameType($use['alias'] ?? null);
                    $name = $this->nameType($use['name']);
                    $this->aliases[$alias] = $name;
                }
                break;
            case 'Stmt_ElseIf':
                $node = new \NodeType\StmtDecisionElseIf();
                break;
            case 'Stmt_Else':
                $node = new \NodeType\StmtDecisionElse();
                break;
            case 'Stmt_Switch':
                $node = new \NodeType\StmtDecisionSwitch();
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
        if ($this->lastStructuredParentStmt
            && method_exists($this->lastStructuredParentStmt, 'setOperators')
            && (strpos($stmt['nodeType'] ?? '', 'Stmt_Expr') !== false || isset($stmt['expr']))
        ) {
            $foundOperators = $this->getOperators($stmt);
            $operators = $this->lastStructuredParentStmt->getOperators() ?? [];
            foreach ($foundOperators as $operator) {
                $operator = new \NodeType\StmtOperator([
                    'name' => $operator,
                ]);
                $operators[] = $operator;
                $this->lastStructuredParentStmt->setOperators($operators);
            }
        }

        // Operands
        if ($this->lastStructuredParentStmt && method_exists($this->lastStructuredParentStmt,
                'setOperators') && ((isset($stmt['expr']['var']) || isset($stmt['var'])))) {
            $name = $this->nameVar($stmt);
            if ($name) {
                // todo utiliser ExprVariable
                $operand = new \NodeType\StmtOperand([
                    'name' => $name,
                ]);
                $operands = $this->lastStructuredParentStmt->getOperands() ?? [];
                $operands[] = $operand;
                $this->lastStructuredParentStmt->setOperands($operands);
            }
        }

        if ($this->lastStructuredParentStmt && method_exists($this->lastStructuredParentStmt, 'getExternals')) {
            // External uses (new, static, etc.)
            $exprs = array_merge($stmt['exprs'] ?? [], (array)($stmt['expr'] ?? []));
            foreach ($exprs as $expr) {
                $usages = $this->lastStructuredParentStmt->getExternals() ?? [];
                if (isset($expr['class'])) {
                    $name = $this->nameType($expr['class']);
                    $usages[] = new \NodeType\Name([
                        'short' => $name,
                        'qualified' => $name,
                    ]);
                }
                $this->lastStructuredParentStmt->setExternals($usages);
            }
        }

        if (!$node) {
            // even if node is not supported, we iterate on its children
            return null;
        }

        // Determine the name if the statement has one
        if (isset($stmt['name'])) {
            $name = $this->nameType($stmt);
            $qualified = $this->lastNamespace . '\\' . $name;
            if ($node instanceof StmtNamespace) {
                $qualified = $name;
            }

            $node->setName(new \NodeType\Name([
                'short' => $name,
                'qualified' => $qualified,
            ]));
        }

        // Extends and implements
        if (isset($stmt['extends'])) {
            $name = $this->nameType($stmt['extends']);
            $extends = $node->getExtends() ?? [];
            $extends[] = new \NodeType\Name([
                'short' => $name,
                'qualified' => $name,
            ]);
            $node->setExtends($extends);
        }

        // Implements
        if (isset($stmt['implements'])) {
            $implements = $node->getImplements() ?? [];
            foreach ($stmt['implements'] as $implement) {
                $name = $this->nameType($implement);
                $implements[] = new \NodeType\Name([
                    'short' => $name,
                    'qualified' => $name,
                ]);
            }
            $node->setImplements($implements);
        }

        // Parameters (for functions and methods)
        if (isset($stmt['params'])) {
            $parameters = $node->getParameters() ?? [];
            foreach ($stmt['params'] as $param) {
                $type = $this->nameType($param['type'] ?? null);
                $parameters[] = new \NodeType\StmtParameter([
                    'name' => $this->nameVar($param),
                    'type' => new \NodeType\Name([
                        'short' => $type,
                        'qualified' => $type,
                    ]),
                ]);
            }
            $node->setParameters($parameters);
        }


        // count blank lines in statement
        if (!empty($stmt['attributes']['startLine'])) {
            $concernedLines = array_slice($this->linesOfCode, $stmt['attributes']['startLine'] - 1,
                $stmt['attributes']['endLine'] - $stmt['attributes']['startLine'] + 1);
            // Location (in code)
            $location = new \NodeType\StmtLocationInFile([
                'startLine' => $stmt['attributes']['startLine'],
                'endLine' => $stmt['attributes']['endLine'],
                'startFilePos' => $stmt['attributes']['startFilePos'],
                'endFilePos' => $stmt['attributes']['endFilePos'],
                'blankLines' => count(array_filter($concernedLines, function ($line) {
                    return trim($line) === '';
                })),
            ]);
            $node->setLocation($location);
        }

        // Count comments lines
        if ($this->lastStructuredParentStmt && empty($this->lastStructuredParentStmt->getLinesOfCode())) {
            $this->lastStructuredParentStmt->setLinesOfCode(new LinesOfCode());
        }

        // Count code lines (for node itself)
        if (method_exists($node, 'getLinesOfCode')) {
            $r = $this->countLinesOfCode($stmt);
            $r['lloc'] = max(1, $r['loc'] - ($r['blanks'] + $r['cloc']));
            $node->setLinesOfCode(new LinesOfCode());
            $node->getLinesOfCode()->setLinesOfCode($r['loc']);
            $node->getLinesOfCode()->setLogicalLinesOfCode($r['lloc']);
            $node->getLinesOfCode()->setCommentLinesOfCode($r['cloc']);
            $node->getLinesOfCode()->setNonCommentLinesOfCode($r['ncloc']);
        }

        return $node;
    }


    public function unalias($name): ?string
    {
        if (isset($this->aliases[$name])) {
            return $this->aliases[$name];
        }

        return $name;
    }

    public function nameVar($what): ?string
    {
        if (isset($what['var']['name'], $what['name']['name'])) {

            // not nested vars
            if (isset($what['var']['var']['name'])) {
                return null;
            }

            $retrieveNameInArrayRecursively = function($item) use(&$retrieveNameInArrayRecursively) {
                $name = $item['name'] ?? null;
                if(is_array($name)) {
                    $name = $retrieveNameInArrayRecursively($name);
                }

                return $name;
            };

            return $retrieveNameInArrayRecursively($what['var']) 
                . '->' 
                . $retrieveNameInArrayRecursively($what['name']);
        }

        if (isset($what['var'])) {
            return $this->nameVar($what['var']);
        }
        if (isset($what['expr'])) {
            return $this->nameVar($what['expr']);
        }

        if(is_array($what['name'])) {
            return $this->nameVar($what['name']);
        }

        return $what['name'] ?? null;
    }

    public function nameType($what): ?string
    {
        if (isset($what['type'])) {
            return $this->nameType($what['type']);
        }

        $name = $what['name'] ?? null;
        $parts = $what['parts'] ?? [];
        if (!empty($parts)) {
            return $this->unalias(implode('\\', $what['parts']));
        }

        if (is_array($name)) {
            $name = $this->nameType($name);
        }

        return $this->unalias($name);
    }


    /**
     * According to the type of statement, we create the corresponding proto node and inject it into the parent node.
     *
     * @param \PhpParser\Node\Stmt $stmt
     * @param Stmt $parent
     * @return \NodeType\StmtClass|\NodeType\StmtFunction|\NodeType\StmtNamespace|null
     */
    private function stmtToProto(array $stmt, \NodeType\Stmts $parent)
    {
        $protoNode = $this->stmtFactory($stmt);
        if (!$protoNode) {
            return null;
        }
        $collection = 'get' . str_replace('NodeType\\', '', get_class($protoNode));
        $parent->$collection()[] = $protoNode;

        // if contains sub statements, do the same for each of them
        $else = $stmt['else'] ?? [];
        if (!is_array($else) || isset($else['nodeType'])) {
            $else = [$else];
        }

        $cases = $stmt['cases'] ?? [];
        if (!is_array($cases) || isset($cases['nodeType'])) {
            $cases = [$cases];
        }

        $elseifs = $stmt['elseifs'] ?? [];
        if (!is_array($elseifs) || isset($elseifs['nodeType'])) {
            $elseifs = [$elseifs];
        }

        $decisions = array_filter(array_merge($else, $cases, $elseifs));
        foreach ($decisions as $decision) {
            $this->stmtToProto($decision, $parent);
        }

        $subStatements = array_filter(array_merge(
            $stmt['stmts'] ?? [],
            [$stmt['stmt'] ?? []]
        ));

        if (!empty($subStatements) && is_array($subStatements)) {
            $protoStmts = new \NodeType\Stmts();
            $protoNode->setStmts($protoStmts);
            foreach ($subStatements as $stmt) {

                if (!is_array($stmt)) {
                    continue;
                }

                // if array is composed only from numeric keys, it's a list of statements
                if (array_keys($stmt) === range(0, count($stmt) - 1)) {
                    foreach ($stmt as $subStmt) {
                        $this->stmtToProto($subStmt, $protoStmts);
                    }
                    continue;
                }

                $this->stmtToProto($stmt, $protoStmts);
            }
        }

        return $protoNode;
    }

    private function countLinesOfCode(array $stmt)
    {
        $result = [
            'loc' => 0,
            'lloc' => 0,
            'cloc' => 0,
            'ncloc' => 0,
            'blanks' => 0,
        ];

        // loc
        if (!empty($stmt['attributes']['startLine'])) {
            $result['loc'] = $stmt['attributes']['endLine'] - $stmt['attributes']['startLine'] + 1;
        }

        if (!empty($stmt['attributes']['comments'])) {
            $stmtsComments = $stmt['attributes']['comments'] ?? [];
            foreach ($stmtsComments as $comment) {
                $result['cloc'] += $comment['endLine'] - $comment['line'] + 1;

                // with php-parser, comments are not included in the loc
                $result['loc'] += $result['cloc'];
            }
        }


        // blank lines
        if (!empty($stmt['attributes']['startLine'])) {
            $concernedLines = array_slice($this->linesOfCode, $stmt['attributes']['startLine'] - 1,
                $stmt['attributes']['endLine'] - $stmt['attributes']['startLine'] + 1);

            $result['blanks'] = count(array_filter($concernedLines, function ($line) {
                return trim($line) === '';
            }));
        }

        // foreach substatement
        foreach ($stmt as $index => $subStmt) {
            // if array is composed only from numeric keys, it's a list of statements
            if (!is_array($subStmt)) {
                continue;
            }


            if (array_keys($subStmt) === range(0, count($subStmt) - 1)) {
                foreach ($subStmt as $subSubStmt) {
                    if (!is_array($subSubStmt)) {
                        continue;
                    }
                    $subResult = $this->countLinesOfCode($subSubStmt);
                    //$result['lloc'] += $subResult['lloc'];
                    $result['cloc'] += $subResult['cloc'];
                }
            }
        }

        // lloc
        //$result['lloc'] = max(1, $result['loc'] - $result['cloc'] - $result['blanks']);

        return $result;

    }

    private function getOperators(array $stmt): array
    {
        $operators = [];
        $expressions = array_merge($stmt['exprs'] ?? [], (array)($stmt['expr'] ?? []));

        foreach ($expressions as $expr) {

            if (!empty($expr['expr'])) {
                $subs = $this->getOperators($expr);
                foreach ($subs as $sub) {
                    $operators[] = $sub;
                }
                continue;
            }

            // if expr is composed only from numeric keys, it's a list of statements
            if (!isset($expr['nodeType']) && is_array($expr)) {
                continue;
            }

            if (is_array($expr) && array_keys($expr) === range(0, count($expr) - 1)) {
                $operators[] = current($expr);
                continue;
            }

            $name = (string) ($expr['nodeType'] ?? $expr);
            if (false === strpos($name, 'Expr_')) {
                continue;
            }

            $name = str_replace('Expr_', '', $name);
            if (in_array($name, ['Variable', 'PropertyFetch', 'MethodCall'])) {
                continue;
            }

            $operators[] = $name;
        }

        return $operators;

    }
}