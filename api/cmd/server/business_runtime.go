package main

import (
	"github.com/gin-gonic/gin"
	"sechelper-auth-template/api/internal/application"
	"sechelper-auth-template/api/internal/modules/authorization"
	"sechelper-auth-template/api/internal/modules/manifest"
)

type businessRuntime struct{ modules *application.Registry }

func (runtime *businessRuntime) SetConfiguration(provider application.ConfigurationProvider) {
	runtime.modules.SetConfiguration(provider)
}

func (runtime *businessRuntime) SetLogger(logger application.Logger) {
	runtime.modules.SetLogger(logger)
}

func (runtime *businessRuntime) RegisterPermissions(registry *manifest.Module) error {
	return runtime.modules.RegisterPermissions(registry)
}

func (runtime *businessRuntime) RegisterResources(registry *authorization.Module) error {
	return runtime.modules.RegisterResources(registry)
}

func (runtime *businessRuntime) RegisterAuditEvents(registry application.AuditRegistrar) error {
	return runtime.modules.RegisterAuditEvents(registry)
}

func (runtime *businessRuntime) RegisterRoutes(v1 *gin.RouterGroup, authorizer application.RouteAuthorizer) error {
	return runtime.modules.RegisterRoutes(v1, authorizer)
}
