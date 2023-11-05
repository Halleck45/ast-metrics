<?php
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: proto/NodeType.proto

namespace GPBMetadata\Proto;

class NodeType
{
    public static $is_initialized = false;

    public static function initOnce() {
        $pool = \Google\Protobuf\Internal\DescriptorPool::getGeneratedPool();

        if (static::$is_initialized == true) {
          return;
        }
        $pool->internalAddGeneratedFile(
            '
�
proto/NodeType.protoNodeType";
Name
short (	
	qualified (	
	describer (	"�
Stmts"
analyze (2.NodeType.Analyze&
	stmtClass (2.NodeType.StmtClass,
stmtFunction (2.NodeType.StmtFunction.
stmtInterface (2.NodeType.StmtInterface&
	stmtTrait (2.NodeType.StmtTrait"
stmtUse (2.NodeType.StmtUse.
stmtNamespace (2.NodeType.StmtNamespace0
stmtDecisionIf (2.NodeType.StmtDecisionIf8
stmtDecisionElseIf	 (2.NodeType.StmtDecisionElseIf4
stmtDecisionElse
 (2.NodeType.StmtDecisionElse4
stmtDecisionCase (2.NodeType.StmtDecisionCase$
stmtLoop (2.NodeType.StmtLoop8
stmtDecisionSwitch (2.NodeType.StmtDecisionSwitch"4
File
path (	
stmts (2.NodeType.Stmts"v
StmtLocationInFile
	startLine (
startFilePos (
endLine (

endFilePos (

blankLines ("}
StmtNamespace
name (2.NodeType.Name
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"w
StmtUse
name (2.NodeType.Name
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"�
	StmtClass
name (2.NodeType.Name
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile\'
comments (2.NodeType.StmtComment)
	operators (2.NodeType.StmtOperator\'
operands (2.NodeType.StmtOperand
extends (2.NodeType.Name"

implements (2.NodeType.Name
uses	 (2.NodeType.Name"�
StmtFunction
name (2.NodeType.Name
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile\'
comments (2.NodeType.StmtComment)
	operators (2.NodeType.StmtOperator\'
operands (2.NodeType.StmtOperand+

parameters (2.NodeType.StmtParameter!
	externals (2.NodeType.Name";
StmtParameter
name (	
type (2.NodeType.Name"�
StmtInterface
name (2.NodeType.Name
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile
extends (2.NodeType.Name"y
	StmtTrait
name (2.NodeType.Name
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"`
StmtDecisionIf
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"d
StmtDecisionElseIf
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"b
StmtDecisionElse
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"b
StmtDecisionCase
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"d
StmtDecisionSwitch
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"Z
StmtLoop
stmts (2.NodeType.Stmts.
location (2.NodeType.StmtLocationInFile"K
StmtComment
text (	.
location (2.NodeType.StmtLocationInFile"
StmtOperator
name (	"
StmtOperand
name (	"�
Analyze(

complexity (2.NodeType.Complexity 
volume (2.NodeType.Volume2
maintainability (2.NodeType.Maintainability"4

Complexity

cyclomatic (H �B
_cyclomatic"�
Volume
loc (H �
lloc (H�
cloc (H�
halsteadVocabulary (H�
halsteadLength (H�
halsteadVolume (H�
halsteadDifficulty (H�
halsteadEffort (H�
halsteadTime	 (H�$
halsteadEstimatedLength
 (H	�B
_locB
_llocB
_clocB
_halsteadVocabularyB
_halsteadLengthB
_halsteadVolumeB
_halsteadDifficultyB
_halsteadEffortB
_halsteadTimeB
_halsteadEstimatedLength"�
Maintainability!
maintainabilityIndex (H �0
#maintainabilityIndexWithoutComments (H�
commentWeight (H�B
_maintainabilityIndexB&
$_maintainabilityIndexWithoutCommentsB
_commentWeightB+Z)github.com/halleck45/ast-metrics/NodeTypebproto3'
        , true);

        static::$is_initialized = true;
    }
}

