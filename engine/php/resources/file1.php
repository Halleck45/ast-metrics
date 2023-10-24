<?php

namespace Foo;

class MyClass1
{
    /**
     * This is a multiline comment
     *
     * @return void
     */
    public function baz1()
    {
        // This is a line comment
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
    #[Attribute]
    public function bar()
    {

    }
}