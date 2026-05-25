// Copyright (c) 2024 Tencent Inc.
// SPDX-License-Identifier: Apache-2.0

package templatecenter

import (
	"os"
	"strings"
	"testing"

	"github.com/tencentcloud/CubeSandbox/CubeMaster/pkg/base/config"
	"github.com/tencentcloud/CubeSandbox/CubeMaster/pkg/service/sandbox/types"
)

func TestMain(m *testing.M) {
	config.SetConfigForTest(&config.Config{
		Scheduler: &config.WrapperSchedulerConf{},
	})
	os.Exit(m.Run())
}

// withConstraints sets ResourceConstraints for the duration of fn, then
// restores the original value. Centralises the save/restore boilerplate.
func withConstraints(t *testing.T, rc *config.ResourceConstraints, fn func(*testing.T)) {
	t.Helper()
	cfg := config.GetConfig()
	orig := cfg.Scheduler.ResourceConstraints
	cfg.Scheduler.ResourceConstraints = rc
	defer func() { cfg.Scheduler.ResourceConstraints = orig }()
	fn(t)
}

func TestValidateResourceConstraints_NilConstraints(t *testing.T) {
	withConstraints(t, nil, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "1m", Mem: "1Mi"}},
		}
		if err := ValidateResourceConstraints(containers); err != nil {
			t.Fatalf("expected nil error when constraints not configured, got: %v", err)
		}
	})
}

func TestValidateResourceConstraints_NilContainers(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		if err := ValidateResourceConstraints(nil); err != nil {
			t.Fatalf("expected nil error for nil containers, got: %v", err)
		}
	})
}

func TestValidateResourceConstraints_ContainerWithNilResources(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: nil},
		}
		if err := ValidateResourceConstraints(containers); err != nil {
			t.Fatalf("expected nil error for nil resources, got: %v", err)
		}
	})
}

func TestValidateResourceConstraints_BelowMinCPU(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "50m", Mem: "256Mi"}},
		}
		err := ValidateResourceConstraints(containers)
		if err == nil {
			t.Fatal("expected error for cpu below minimum")
		}
		if !strings.Contains(err.Error(), "cpu '50m' is below minimum '100m'") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestValidateResourceConstraints_BelowMinMemory(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "2000m", Mem: "64Mi"}},
		}
		err := ValidateResourceConstraints(containers)
		if err == nil {
			t.Fatal("expected error for memory below minimum")
		}
		if !strings.Contains(err.Error(), "memory '64Mi' is below minimum '128Mi'") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestValidateResourceConstraints_AtExactMinimum(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "100m", Mem: "128Mi"}},
		}
		if err := ValidateResourceConstraints(containers); err != nil {
			t.Fatalf("expected nil error at exact minimum, got: %v", err)
		}
	})
}

func TestValidateResourceConstraints_AboveMinimum(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "2000m", Mem: "2000Mi"}},
		}
		if err := ValidateResourceConstraints(containers); err != nil {
			t.Fatalf("expected nil error above minimum, got: %v", err)
		}
	})
}

func TestValidateResourceConstraints_MultipleContainers(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "50m", Mem: "64Mi"}},
			{Name: "c2", Resources: &types.Resource{Cpu: "50m", Mem: "256Mi"}},
		}
		err := ValidateResourceConstraints(containers)
		if err == nil {
			t.Fatal("expected error for containers below minimum")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "container 'c1'") {
			t.Fatalf("expected error to reference container 'c1', got: %v", err)
		}
		if !strings.Contains(errStr, "container 'c2'") {
			t.Fatalf("expected error to reference container 'c2', got: %v", err)
		}
	})
}

func TestValidateResourceConstraints_BothCPUAndMemoryViolation(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "50m", Mem: "64Mi"}},
		}
		err := ValidateResourceConstraints(containers)
		if err == nil {
			t.Fatal("expected error for both cpu and memory below minimum")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "cpu '50m'") {
			t.Fatalf("expected cpu violation in error, got: %v", err)
		}
		if !strings.Contains(errStr, "memory '64Mi'") {
			t.Fatalf("expected memory violation in error, got: %v", err)
		}
	})
}

func TestValidateResourceConstraints_InvalidCPUString(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "invalid", Mem: "256Mi"}},
		}
		err := ValidateResourceConstraints(containers)
		if err == nil {
			t.Fatal("expected error for invalid cpu value")
		}
		if !strings.Contains(err.Error(), "invalid cpu value") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestValidateResourceConstraints_InvalidMemoryString(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "200m", Mem: "invalid"}},
		}
		err := ValidateResourceConstraints(containers)
		if err == nil {
			t.Fatal("expected error for invalid memory value")
		}
		if !strings.Contains(err.Error(), "invalid memory value") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestValidateResourceConstraints_EmptyResourceValues(t *testing.T) {
	rc := &config.ResourceConstraints{MinCPU: "100m", MinMemory: "128Mi"}
	withConstraints(t, rc, func(t *testing.T) {
		containers := []*types.Container{
			{Name: "c1", Resources: &types.Resource{Cpu: "", Mem: ""}},
		}
		if err := ValidateResourceConstraints(containers); err != nil {
			t.Fatalf("expected nil error for empty resource values, got: %v", err)
		}
	})
}
