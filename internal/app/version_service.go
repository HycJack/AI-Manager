package app

import (
	"ai-manager/internal/version"
)

// VersionService exposes build-time metadata to the frontend.
type VersionService struct{}

// NewVersionService creates the version service.
func NewVersionService() *VersionService { return &VersionService{} }

// GetInfo returns the version, commit hash, and build time.
func (s *VersionService) GetInfo() version.Info {
	return version.Get()
}
