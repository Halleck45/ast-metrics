<?php
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: proto/NodeType.proto

namespace NodeType;

use Google\Protobuf\Internal\GPBType;
use Google\Protobuf\Internal\RepeatedField;
use Google\Protobuf\Internal\GPBUtil;

/**
 * Generated from protobuf message <code>NodeType.Volume</code>
 */
class Volume extends \Google\Protobuf\Internal\Message
{
    /**
     * Generated from protobuf field <code>optional int32 loc = 1;</code>
     */
    protected $loc = null;
    /**
     * Generated from protobuf field <code>optional int32 lloc = 2;</code>
     */
    protected $lloc = null;
    /**
     * Generated from protobuf field <code>optional int32 cloc = 3;</code>
     */
    protected $cloc = null;

    /**
     * Constructor.
     *
     * @param array $data {
     *     Optional. Data for populating the Message object.
     *
     *     @type int $loc
     *     @type int $lloc
     *     @type int $cloc
     * }
     */
    public function __construct($data = NULL) {
        \GPBMetadata\Proto\NodeType::initOnce();
        parent::__construct($data);
    }

    /**
     * Generated from protobuf field <code>optional int32 loc = 1;</code>
     * @return int
     */
    public function getLoc()
    {
        return isset($this->loc) ? $this->loc : 0;
    }

    public function hasLoc()
    {
        return isset($this->loc);
    }

    public function clearLoc()
    {
        unset($this->loc);
    }

    /**
     * Generated from protobuf field <code>optional int32 loc = 1;</code>
     * @param int $var
     * @return $this
     */
    public function setLoc($var)
    {
        GPBUtil::checkInt32($var);
        $this->loc = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>optional int32 lloc = 2;</code>
     * @return int
     */
    public function getLloc()
    {
        return isset($this->lloc) ? $this->lloc : 0;
    }

    public function hasLloc()
    {
        return isset($this->lloc);
    }

    public function clearLloc()
    {
        unset($this->lloc);
    }

    /**
     * Generated from protobuf field <code>optional int32 lloc = 2;</code>
     * @param int $var
     * @return $this
     */
    public function setLloc($var)
    {
        GPBUtil::checkInt32($var);
        $this->lloc = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>optional int32 cloc = 3;</code>
     * @return int
     */
    public function getCloc()
    {
        return isset($this->cloc) ? $this->cloc : 0;
    }

    public function hasCloc()
    {
        return isset($this->cloc);
    }

    public function clearCloc()
    {
        unset($this->cloc);
    }

    /**
     * Generated from protobuf field <code>optional int32 cloc = 3;</code>
     * @param int $var
     * @return $this
     */
    public function setCloc($var)
    {
        GPBUtil::checkInt32($var);
        $this->cloc = $var;

        return $this;
    }

}
