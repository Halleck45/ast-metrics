<?php

namespace Foo;

use SplObjectStorage as MyAlias;

class MyClass1 extends \Bar\MyClass2 implements \JsonSerializable
{
    /**
     * This is a multiline comment
     *
     * @return void
     */
    public function baz1($eVar, int $dVar, \MyAlias $storageVar = null)
    {
        $aVar = 1;
        $bVar = 2;
        $cVar = 3;

        $dVar = new \LogicException('Hello World');


        // This is a line comment
        if (true) {
            echo "Hello";
            if (false) {
                echo "World";
            }

            foreach (array(1, 2, 3) as $value) {
                echo $value;

                $aVar = 4;
            }

        } elseif (false) {
            echo "World";
        } else {
            echo $cVar;
        }

        $bVar++;
    }

    public function jsonSerialize()
    {
        return [];
    }
}

class MyClass2
{
    #[Attribute]
    public function bar()
    {

    }
}