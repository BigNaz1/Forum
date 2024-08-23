package RebootForums

import (
	"sync"
)

var (
	templatesDir string
	templatesMu  sync.RWMutex
)

// SetTemplatesDir sets the directory for HTML templates
func SetTemplatesDir(dir string) {
	templatesMu.Lock()
	templatesDir = dir
	templatesMu.Unlock()
}

// GetTemplatesDir returns the directory for HTML templates
func GetTemplatesDir() string {
	templatesMu.RLock()
	defer templatesMu.RUnlock()
	return templatesDir
}
