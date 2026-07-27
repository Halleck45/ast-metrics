package ruleset

import (
	"github.com/ast-metrics/ast-metrics/internal/analyzer/issue"
	pb "github.com/ast-metrics/ast-metrics/pb"
)

type Rule interface {
	Name() string
	Description() string
	CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string))
}
