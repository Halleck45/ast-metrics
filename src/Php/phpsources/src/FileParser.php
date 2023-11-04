<?php

namespace App;

class FileParser {

    public function parse(string $file): array
    {
        $lexer = new \PhpParser\Lexer\Emulative(['usedAttributes' => [
            'startLine', 'endLine', 'startFilePos', 'endFilePos', 'comments'
        ]]);
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
}