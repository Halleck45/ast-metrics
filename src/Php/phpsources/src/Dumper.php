#!/usr/bin/env php
<?php

// The code is directly inspired from nikic/php-parser/bin/php-parse
//
// The main difference is that we don't dump the JSON as it is, but we standardize it with our protobuf schemas.

namespace App;

use Google\Protobuf\Internal\RepeatedField;
use NodeType\File;
use NodeType\StmtNamespace;


class Dumper
{
    private ?RepeatedField $comments;
    private array $aliases = [];
    private string $file;
    private $lastStructuredParentStmt;
    private string $lastNamespace;
    private array $linesOfCode=[];

    public function __construct(string $file)
    {
        $this->file = $file;
    }

    public function dump(array $json): File
    {

        // This main node describe the file itself
        $fileNode = new File([
            'path' => realpath($this->file),
        ]);
        $protoStmts = new \NodeType\Stmts();
        $fileNode->setStmts($protoStmts);

        $subs = [];
        if (getenv('DEBUG')) {
            file_put_contents('tmp.json', json_encode($json, JSON_PRETTY_PRINT));
        }
        foreach ($json as $stmt) {
            $this->stmtToProto($stmt, $protoStmts);
        }

        return $fileNode;
    }
    
    
    private function stmtFactory(array $stmt)
    {
        $node = null;
        switch ($stmt['nodeType'] ?? null) {
            case 'Stmt_Namespace':
                $node = new \NodeType\StmtNamespace();
                $this->lastNamespace = $this->nameType($stmt['name']);
                $this->aliases = [];
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
            && (strpos($stmt['nodeType'] ?? '', 'Stmt_Expr') !== false || isset($stmt['expr']))
        ) {
            $operator = new \NodeType\StmtOperator([
                'name' => $stmt['expr']['nodeType'] ?? $stmt['nodeType'] ?? $stmt['expr'] ?? null,
            ]);
            $operators = $this->lastStructuredParentStmt->getOperators() ?? [];
            $operators[] = $operator;
            $this->lastStructuredParentStmt->setOperators($operators);
        }
        // Operands
        if ($this->lastStructuredParentStmt && ((isset($stmt['expr']['var']) || isset($stmt['var'])))) {
            $name = $this->nameVar($stmt);
            if ($name) {
                $operand = new \NodeType\StmtOperand([
                    'name' => $name,
                ]);
                $operands = $this->lastStructuredParentStmt->getOperands() ?? [];
                $operands[] = $operand;
                $this->lastStructuredParentStmt->setOperands($operands);
            }

        }

        if ($this->lastStructuredParentStmt) {
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
            return null;
        }

        // Determine the name if the statement has one
        if (isset($stmt['name'])) {
            $name = $this->nameType($stmt);
            $qualified = $this->lastNamespace . $name;
            if($node instanceof StmtNamespace) {
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
        $concernedLines = array_slice($this->linesOfCode, $stmt['attributes']['startLine'] - 1, $stmt['attributes']['endLine'] - $stmt['attributes']['startLine'] + 1);
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

        // Determine if the statement is a decision or a structure
        if (method_exists($node, 'setComments')) {
            $this->lastStructuredParentStmt = $node;
        }


        if (!empty($stmt['attributes']['comments'])) {
            if ($this->lastStructuredParentStmt) {
                // Node is a class or a method
                $this->comments = $this->lastStructuredParentStmt->getComments() ?? [];
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
                    $this->comments[] = $protoComment;
                }

                $this->lastStructuredParentStmt->setComments($this->comments);
            }
        }

        if (!empty($stmt['attributes']['comments'])) {
            if ($this->lastStructuredParentStmt) {
                // Node is a class or a method
                $this->comments = $this->lastStructuredParentStmt->getComments() ?? [];
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
                    $this->comments[] = $protoComment;
                }

                $this->lastStructuredParentStmt->setComments($this->comments);
            }
        }

        return $node;
    }


    public function unalias($name) : ?string {
        if (isset($this->aliases[$name])) {
            return $this->aliases[$name];
        }

        return $name;
    }


    public function nameVar($what): ?string
    {
        if (isset($what['var'])) {
            return $this->nameVar($what['var']);
        }
        if (isset($what['expr'])) {
            return $this->nameVar($what['expr']);
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
        $subStatements = array_filter(array_merge(
            $stmt['stmts'] ?? [],
            $stmt['cases'] ?? [],
            $stmt['else'] ?? [],
            $stmt['elseifs'] ?? [],
            [$stmt['stmt'] ?? []]
        ));
        if (!empty($subStatements) && is_array($subStatements)) {
            $protoStmts = new \NodeType\Stmts();
            $protoNode->setStmts($protoStmts);
            foreach ($subStatements as $stmt) {
                if (!is_array($stmt)) {
                    continue;
                }

                $this->stmtToProto($stmt, $protoStmts);
            }
        }

        return $protoNode;
    }
}