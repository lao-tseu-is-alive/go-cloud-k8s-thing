package module

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"connectrpc.com/vanguard"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1/thingv1connect"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/thing"
)

// RoutePattern describes a REST route pattern to mount on Echo.
type RoutePattern struct {
	Pattern string // e.g. "/thing*", "/types*"
}

// VanguardServices returns the Vanguard services for this module.
//
// In bundle mode, the caller collects services from all modules and
// creates a single vanguard.NewTranscoder for all of them.
func (m *Module) VanguardServices() []*vanguard.Service {
	// Interceptors (JWT + validate)
	authInterceptor := thing.NewAuthInterceptor(m.deps.JWT, m.deps.Logger)
	interceptors := connect.WithInterceptors(authInterceptor, validate.NewInterceptor())

	// Connect servers
	thingConnectServer := thing.NewThingConnectServer(m.business, m.deps.Logger)
	typeThingConnectServer := thing.NewTypeThingConnectServer(m.business, m.deps.Logger)

	// Service handlers
	_, thingHandler := thingv1connect.NewThingServiceHandler(thingConnectServer, interceptors)
	_, typeThingHandler := thingv1connect.NewTypeThingServiceHandler(typeThingConnectServer, interceptors)

	return []*vanguard.Service{
		vanguard.NewService(thingv1connect.ThingServiceName, thingHandler),
		vanguard.NewService(thingv1connect.TypeThingServiceName, typeThingHandler),
	}
}

// RoutePatterns returns the REST route patterns that this module
// handles under the secured prefix (e.g. /goapi/v1/thing*).
func (m *Module) RoutePatterns() []RoutePattern {
	return []RoutePattern{
		{Pattern: "/thing*"},
		{Pattern: "/types*"},
	}
}

// ConnectPatterns returns the gRPC/Connect route patterns that this
// module handles (without prefix).
func (m *Module) ConnectPatterns() []string {
	return []string{"/thing.v1.*"}
}

// RegisterRoutes is a convenience method for standalone mode.
//
// It creates a Vanguard transcoder from VanguardServices() and mounts
// all route patterns on the given Echo instance. For bundle mode, use
// VanguardServices() + RoutePatterns() + ConnectPatterns() instead so
// that a single shared transcoder is used across all modules.
func (m *Module) RegisterRoutes(e *echo.Echo) error {
	services := m.VanguardServices()

	transcoder, err := vanguard.NewTranscoder(services)
	if err != nil {
		return err
	}

	prefix := m.cfg.SecuredPrefix
	transcoderWithPrefix := http.StripPrefix(prefix, transcoder)

	for _, p := range m.RoutePatterns() {
		e.Any(prefix+p.Pattern, echo.WrapHandler(transcoderWithPrefix))
	}

	for _, p := range m.ConnectPatterns() {
		e.Any(p, echo.WrapHandler(transcoder))
	}

	m.deps.Logger.Info("Thing module routes registered", "prefix", prefix)
	return nil
}
