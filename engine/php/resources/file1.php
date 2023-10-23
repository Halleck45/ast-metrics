<?php

namespace Foo;

class MyClass1
{
    public function baz1()
    {
        if (true) {
            echo "Hello";
            if (false) {
                echo "World";
            }

            foreach (array(1, 2, 3) as $value) {
                echo $value;
            }

        } elseif (false) {
            echo "World";
        } else {
            echo "World";
        }
    }
}

class MyClass2
{
    public function bar()
    {

    }
}