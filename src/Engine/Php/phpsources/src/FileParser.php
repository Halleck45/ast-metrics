<?php

namespace App;

class FileParser
{

    private bool $withDetails = true;

    public function parse(string $file): array
    {
        $attributes = [];
        if ($this->withDetails) {
            $attributes = [
                'startLine',
                'endLine',
                'startFilePos',
                'endFilePos',
                'comments'
            ];
        }

        $lexer = new \PhpParser\Lexer\Emulative(['usedAttributes' => $attributes]);
        $parser = (new \PhpParser\ParserFactory)->create(
            \PhpParser\ParserFactory::PREFER_PHP7,
            $lexer
        );
        $traverser = new \PhpParser\NodeTraverser();
        $traverser->addVisitor(new \PhpParser\NodeVisitor\NameResolver);

        $code = file_get_contents($file);
        $stmts = $parser->parse($code);
        return (array)json_decode(json_encode($stmts), true);
    }

    public function enableDetails(bool $withDetails = true)
    {
        $this->withDetails = $withDetails;
    }
}