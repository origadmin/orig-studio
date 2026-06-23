package migrate

import (
	"context"
)

func init() {
	RegisterPhase(PhaseDef{Name: PhaseAds, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseAds(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseTenants, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseTenants(ctx)
	}})
}

const (
	PhaseAds     Phase = "ads"
	PhaseTenants Phase = "tenants"
)

type AdRecord struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ImageURL  string `json:"image_url"`
	LinkURL   string `json:"link_url"`
	IsActive  bool   `json:"is_active"`
	Placement string `json:"placement"`
	Priority  int    `json:"priority"`
}

type TenantRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Domain  string `json:"domain"`
	Plan    string `json:"plan"`
	MaxUsers int   `json:"max_users"`
}

func (e *Engine) phaseAds(ctx context.Context) error {
	e.logger.Printf("Ads migration phase (EE-only) - placeholder")
	e.state.PhaseState[string(PhaseAds)] = string(StatusCompleted)
	return nil
}

func (e *Engine) phaseTenants(ctx context.Context) error {
	e.logger.Printf("Tenants migration phase (EE-only) - placeholder")
	e.state.PhaseState[string(PhaseTenants)] = string(StatusCompleted)
	return nil
}
