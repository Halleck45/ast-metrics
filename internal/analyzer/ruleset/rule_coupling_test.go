package ruleset

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

func TestCouplingRule_Name(t *testing.T) {
	rule := NewCouplingRule(nil)
	if rule.Name() != "coupling" {
		t.Errorf("expected 'coupling', got %s", rule.Name())
	}
}

func TestCouplingRule_Description(t *testing.T) {
	rule := NewCouplingRule(nil)
	expected := "Checks for forbidden coupling between packages"
	if rule.Description() != expected {
		t.Errorf("expected '%s', got %s", expected, rule.Description())
	}
}

func TestCouplingRule_CheckFile_NilConfig(t *testing.T) {
	rule := NewCouplingRule(nil)
	file := &pb.File{
		Path: "/src/controller/UserController.go",
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{ClassName: "repository.UserRepository"},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 0 || len(successes) != 0 {
		t.Error("expected no errors or successes with nil config")
	}
}

func TestCouplingRule_CheckFile_ForbiddenCoupling(t *testing.T) {
	cfg := &configuration.ConfigurationCouplingRule{
		Forbidden: []struct {
			From string `yaml:"from"`
			To   string `yaml:"to"`
		}{
			{From: "Controller", To: "Repository"},
		},
	}
	rule := NewCouplingRule(cfg)
	
	file := &pb.File{
		Path: "/src/controller/UserController.go",
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{ClassName: "repository.UserRepository"},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	err := errors[0]
	if err.Code != "coupling" {
		t.Errorf("expected 'coupling' code, got %s", err.Code)
	}
	if err.Severity != issue.SeverityUnknown {
		t.Errorf("expected unknown severity, got %s", err.Severity)
	}
}

func TestCouplingRule_CheckFile_AllowedCoupling(t *testing.T) {
	cfg := &configuration.ConfigurationCouplingRule{
		Forbidden: []struct {
			From string `yaml:"from"`
			To   string `yaml:"to"`
		}{
			{From: "Controller", To: "Repository"},
		},
	}
	rule := NewCouplingRule(cfg)
	
	file := &pb.File{
		Path: "/src/controller/UserController.go",
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{ClassName: "service.UserService"},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}
	if len(successes) != 1 {
		t.Fatalf("expected 1 success, got %d", len(successes))
	}
	if successes[0] != "Coupling OK" {
		t.Errorf("unexpected success message: %s", successes[0])
	}
}

func TestCouplingRule_CheckFile_NoMatchingFromPattern(t *testing.T) {
	cfg := &configuration.ConfigurationCouplingRule{
		Forbidden: []struct {
			From string `yaml:"from"`
			To   string `yaml:"to"`
		}{
			{From: "Controller", To: "Repository"},
		},
	}
	rule := NewCouplingRule(cfg)
	
	file := &pb.File{
		Path: "/src/service/UserService.go",
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{ClassName: "repository.UserRepository"},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	if len(errors) != 0 {
		t.Errorf("expected no errors when from pattern doesn't match, got %d", len(errors))
	}
	if len(successes) != 1 {
		t.Fatalf("expected 1 success, got %d", len(successes))
	}
}

func TestCouplingRule_CheckFile_MultipleForbiddenRules(t *testing.T) {
	cfg := &configuration.ConfigurationCouplingRule{
		Forbidden: []struct {
			From string `yaml:"from"`
			To   string `yaml:"to"`
		}{
			{From: "Controller", To: "Repository"},
			{From: "Controller", To: "Database"},
		},
	}
	rule := NewCouplingRule(cfg)
	
	file := &pb.File{
		Path: "/src/controller/UserController.go",
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{ClassName: "repository.UserRepository"},
				{ClassName: "database.Connection"},
			},
		},
	}

	errors := []issue.RequirementError{}
	successes := []string{}

	rule.CheckFile(file, 
		func(e issue.RequirementError) { errors = append(errors, e) },
		func(s string) { successes = append(successes, s) })

	// Should stop at first violation
	if len(errors) != 1 {
		t.Fatalf("expected 1 error (stops at first violation), got %d", len(errors))
	}
}
