package server

import "github.com/gin-gonic/gin"

// RouteRegister is an interface for modules that register routes.
type RouteRegister interface {
	Register(router *gin.RouterGroup)
}
