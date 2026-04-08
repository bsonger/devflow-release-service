package model

import (
	"testing"
	"time"
)

func TestBuildImageTargetMainBranchKeepsBaseName(t *testing.T) {
	target, err := BuildImageTarget(ImageRegistryConfig{
		Registry:  "registry.cn-hangzhou.aliyuncs.com",
		Namespace: "devflow",
	}, "Portal API", "main", "", time.Date(2026, 4, 8, 13, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BuildImageTarget returned error: %v", err)
	}
	if target.Name != "portal-api" {
		t.Fatalf("Name = %q want portal-api", target.Name)
	}
	if target.Tag != "20260408-130500" {
		t.Fatalf("Tag = %q want 20260408-130500", target.Tag)
	}
	if target.Ref != "registry.cn-hangzhou.aliyuncs.com/devflow/portal-api:20260408-130500" {
		t.Fatalf("Ref = %q", target.Ref)
	}
}

func TestBuildImageTargetFeatureBranchAppendsNormalizedBranch(t *testing.T) {
	target, err := BuildImageTarget(ImageRegistryConfig{
		Registry:  "registry.cn-hangzhou.aliyuncs.com",
		Namespace: "devflow",
	}, "Portal API", "feature/login", "", time.Date(2026, 4, 8, 13, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BuildImageTarget returned error: %v", err)
	}
	if target.Name != "portal-api-feature-login" {
		t.Fatalf("Name = %q want portal-api-feature-login", target.Name)
	}
	if target.Ref != "registry.cn-hangzhou.aliyuncs.com/devflow/portal-api-feature-login:20260408-130500" {
		t.Fatalf("Ref = %q", target.Ref)
	}
}

func TestBuildImageTargetPrefersGitTag(t *testing.T) {
	target, err := BuildImageTarget(ImageRegistryConfig{
		Registry:  "registry.cn-hangzhou.aliyuncs.com",
		Namespace: "devflow",
	}, "Portal API", "release/v1", "v1.2.3", time.Date(2026, 4, 8, 13, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BuildImageTarget returned error: %v", err)
	}
	if target.Tag != "v1.2.3" {
		t.Fatalf("Tag = %q want v1.2.3", target.Tag)
	}
	if target.Ref != "registry.cn-hangzhou.aliyuncs.com/devflow/portal-api-release-v1:v1.2.3" {
		t.Fatalf("Ref = %q", target.Ref)
	}
}

func TestBuildImageTargetRejectsEmptyApplicationName(t *testing.T) {
	_, err := BuildImageTarget(ImageRegistryConfig{
		Registry:  "registry.cn-hangzhou.aliyuncs.com",
		Namespace: "devflow",
	}, "___", "main", "", time.Date(2026, 4, 8, 13, 5, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected error for empty normalized application name")
	}
}
