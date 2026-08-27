package migrate

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"origadmin/application/origstudio/internal/conf"
)

// SourceAdapterFactory builds a source adapter for one platform. The factory
// receives the target storage layout so adapters can map source-relative file
// paths onto target-relative storage keys (same as MediaCMSAdapter).
type SourceAdapterFactory func(paths *conf.StoragePaths) SourceAdapter

var (
	adapterRegistryMu sync.RWMutex
	adapterRegistry   = map[string]SourceAdapterFactory{}
)

// RegisterSource registers a source platform adapter factory under typeName.
// Adapters self-register in their own file via init(); adding a new platform
// is purely additive (new file + RegisterSource call), no engine or CLI change.
func RegisterSource(typeName string, factory SourceAdapterFactory) {
	adapterRegistryMu.Lock()
	defer adapterRegistryMu.Unlock()
	if _, dup := adapterRegistry[typeName]; dup {
		panic(fmt.Sprintf("migrate: source adapter %q already registered", typeName))
	}
	adapterRegistry[typeName] = factory
}

// NewSourceAdapter returns the adapter for typeName, or an error listing the
// registered platforms when the type is unknown.
func NewSourceAdapter(typeName string, paths *conf.StoragePaths) (SourceAdapter, error) {
	adapterRegistryMu.RLock()
	factory, ok := adapterRegistry[typeName]
	adapterRegistryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown source type %q (available: %s)",
			typeName, strings.Join(RegisteredSources(), ", "))
	}
	return factory(paths), nil
}

// RegisteredSources returns the sorted list of registered source platform types.
func RegisteredSources() []string {
	adapterRegistryMu.RLock()
	defer adapterRegistryMu.RUnlock()
	names := make([]string, 0, len(adapterRegistry))
	for name := range adapterRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
