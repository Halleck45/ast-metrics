package Configuration

import (
    "github.com/halleck45/ast-metrics/src/Driver"
    "testing"
)


func TestConfigurationAcceptsDriver(t *testing.T) {

    configuration := NewConfiguration()
    configuration.SetDriver(Driver.Docker)

    if configuration.Driver != Driver.Docker {
        t.Errorf("Driver = %s; want %s", configuration.Driver, Driver.Docker)
    }
}

func TestConfigurationAcceptsSourcesToAnalyzePath(t *testing.T) {

    configuration := NewConfiguration()
    configuration.SetSourcesToAnalyzePath([]string{"/foo"})

    if configuration.SourcesToAnalyzePath[0] != "/foo" {
        t.Errorf("SourcesToAnalyzePath = %s; want %s", configuration.SourcesToAnalyzePath[0], "/foo")
    }
}

func TestConfigurationAcceptsExcludePatterns(t *testing.T) {

    configuration := NewConfiguration()
    configuration.SetExcludePatterns([]string{"/foo"})

    if configuration.ExcludePatterns[0] != "/foo" {
        t.Errorf("ExcludePatterns = %s; want %s", configuration.ExcludePatterns[0], "/foo")
    }
}