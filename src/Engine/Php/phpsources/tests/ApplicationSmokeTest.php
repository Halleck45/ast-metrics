<?php

namespace App;

use PHPUnit\Framework\TestCase;

class ApplicationSmokeTest extends TestCase
{

    public function testApplicationDumpsAstAsJson()
    {
        $dir = __DIR__;
        $command = "OUTPUT_FORMAT=json php $dir/../dump.php $dir/resources/smoke1.php";
        $output = shell_exec($command);
        $this->assertJson($output);

        $json = json_decode($output, true);

        $this->assertEquals(realpath($dir . '/resources/smoke1.php'), $json['path']);
        $this->assertEquals("Foo\\Ns1", $json['stmts']['stmtNamespace'][0]['name']['qualified']);
    }
}