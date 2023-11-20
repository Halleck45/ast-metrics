#!/usr/bin/env php
<?php

// Only for debugging
// display JSON from binary file
require_once __DIR__ . '/vendor/autoload.php';
spl_autoload_register(function ($class) {
    $class = str_replace('\\', '/', $class);
    require_once __DIR__ . "/generated/$class.php";
});

if(!isset($argv[1])) {
    echo "Usage:  php read.php <binary-file>\n";
    exit(1);
}
$file = (string) $argv[1];

// load the file using protobuf
$fileNode = new \NodeType\File;
$fileNode->mergeFromString(file_get_contents($file));
echo json_encode(json_decode($fileNode->serializeToJsonString()), JSON_PRETTY_PRINT);
