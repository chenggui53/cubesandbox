// Copyright (c) 2024 Tencent Inc.
// SPDX-License-Identifier: Apache-2.0

package templatecenter

import (
	"fmt"
	"strings"

	"github.com/tencentcloud/CubeSandbox/CubeMaster/pkg/base/config"
	"github.com/tencentcloud/CubeSandbox/CubeMaster/pkg/service/sandbox/types"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ValidateResourceConstraints checks that container resources meet minimum
// requirements defined in the scheduler config. Returns nil if no constraints
// are configured or if all containers pass validation. All violations across
// containers are collected into a single error.
func ValidateResourceConstraints(containers []*types.Container) error {
	cfg := config.GetConfig()
	if cfg == nil || cfg.Scheduler == nil || cfg.Scheduler.ResourceConstraints == nil {
		return nil
	}
	rc := cfg.Scheduler.ResourceConstraints

	var errs []string
	for _, ctr := range containers {
		if ctr == nil || ctr.Resources == nil {
			continue
		}
		if err := validateContainerResource(ctr.Name, ctr.Resources, rc); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("resource constraint violations:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

func validateContainerResource(name string, res *types.Resource, rc *config.ResourceConstraints) error {
	var errs []string

	if res.Cpu != "" && rc.MinCPU != "" {
		cpu, err := resource.ParseQuantity(res.Cpu)
		if err != nil {
			return fmt.Errorf("container '%s' has invalid cpu value '%s': %w", name, res.Cpu, err)
		}
		minCPU, err := rc.MinCPURes()
		if err != nil {
			return fmt.Errorf("container '%s': %w", name, err)
		}
		if cpu.Cmp(minCPU) < 0 {
			errs = append(errs, fmt.Sprintf("container '%s' cpu '%s' is below minimum '%s'", name, res.Cpu, rc.MinCPU))
		}
	}
	if res.Mem != "" && rc.MinMemory != "" {
		mem, err := resource.ParseQuantity(res.Mem)
		if err != nil {
			return fmt.Errorf("container '%s' has invalid memory value '%s': %w", name, res.Mem, err)
		}
		minMem, err := rc.MinMemoryRes()
		if err != nil {
			return fmt.Errorf("container '%s': %w", name, err)
		}
		if mem.Cmp(minMem) < 0 {
			errs = append(errs, fmt.Sprintf("container '%s' memory '%s' is below minimum '%s'", name, res.Mem, rc.MinMemory))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}
