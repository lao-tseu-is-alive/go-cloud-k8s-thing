// Package module provides an importable Module for the Thing domain.
//
// It encapsulates the domain wiring (storage + business service) and
// transport wiring (Connect + Vanguard) so that the same code can be
// used both in the standalone go-cloud-k8s-thing server and in a
// multi-service bundle.
package module

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/thing"
)

// Deps holds the cross-cutting dependencies injected by the main (or bundle).
type Deps struct {
	DB     database.DB
	JWT    goHttpEcho.JwtChecker // already an interface in goHttpEcho
	Logger *slog.Logger
}

// Config holds module-specific configuration.
type Config struct {
	SecuredPrefix    string // e.g. "/goapi/v1"
	ListDefaultLimit int    // e.g. 50
}

// Module encapsulates the Thing domain: storage, business service,
// and transport handlers.
type Module struct {
	cfg  Config
	deps Deps

	business *thing.BusinessService
}

// New creates a new Thing Module. The provided ctx is used for the
// initial storage verification query (e.g. checking that reference
// tables are populated).
func New(ctx context.Context, cfg Config, deps Deps) (*Module, error) {
	store, err := thing.GetStorageInstance(ctx, "pgx", deps.DB, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("thing module: storage init failed: %w", err)
	}
	bs := thing.NewBusinessService(store, deps.DB, deps.Logger, cfg.ListDefaultLimit)
	return &Module{cfg: cfg, deps: deps, business: bs}, nil
}

// Start is a placeholder for future background workers (e.g. JetStream consumers).
// Currently a no-op.
func (m *Module) Start(_ context.Context) error { return nil }

// Stop is a placeholder for graceful shutdown of background workers.
// Currently a no-op.
func (m *Module) Stop(_ context.Context) error { return nil }
