#!/usr/bin/env php
<?php

require_once __DIR__ . '/vendor/autoload.php';

ini_set('xdebug.max_nesting_level', 3000);
ini_set('xdebug.var_display_max_children', -1);
ini_set('xdebug.var_display_max_data', -1);
ini_set('xdebug.var_display_max_depth', -1);

if(!isset($argv[1])) {
    echo "Usage: DEBUG=1 OUTPUT_FORMAT=json php dump.php <file>\n";
    exit(1);
}
$file = $argv[1];

if (!file_exists($file)) {
    fwrite(STDERR, "File $file does not exist.\n");
    exit(1);
}

if (!is_readable($file)) {
    fwrite(STDERR, "File $file is not readable.\n");
    exit(1);
}

$parser = new \App\FileParser();
$json = $parser->parse($file);

$node = new \App\Dumper($file);
$protoFile = $node->dump($json);

$format = getenv('OUTPUT_FORMAT') ?: 'binary';
switch($format) {

    case 'json-pretty':
        echo json_encode(json_decode($protoFile->serializeToJsonString()), JSON_PRETTY_PRINT);
        break;
    case 'binary':
        echo $protoFile->serializeToString();
        break;
    case 'raw':
        echo json_encode($json, JSON_PRETTY_PRINT);
        break;
    case 'null':
        break;
    case 'json':
    default:
    echo $protoFile->serializeToJsonString();
    break;
}