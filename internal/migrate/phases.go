package migrate

import (
	"context"
)

type PhaseFunc func(ctx context.Context, e *Engine) error

type PhaseDef struct {
	Name Phase
	Fn   PhaseFunc
}

var defaultPhases []PhaseDef

func RegisterPhase(def PhaseDef) {
	defaultPhases = append(defaultPhases, def)
}

func DefaultPhases() []PhaseDef {
	result := make([]PhaseDef, len(defaultPhases))
	copy(result, defaultPhases)
	return result
}

func init() {
	RegisterPhase(PhaseDef{Name: PhaseDiscover, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseDiscover(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseUsers, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseUsers(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseCategories, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseCategories(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseTags, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseTags(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseChannels, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseChannels(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseMedia, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseMedia(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseComments, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseComments(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhasePlaylists, Fn: func(ctx context.Context, e *Engine) error {
		return e.phasePlaylists(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseFiles, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseFiles(ctx)
	}})
	RegisterPhase(PhaseDef{Name: PhaseVerify, Fn: func(ctx context.Context, e *Engine) error {
		return e.phaseVerify(ctx)
	}})
}
