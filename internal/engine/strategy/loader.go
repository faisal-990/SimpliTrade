package strategy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Kind is the evaluator family a strategy maps to. The 9 declared styles collapse
// into 3 evaluators that share pricing/sizing helpers.
type Kind string

const (
	KindFundamental Kind = "fundamental" // value, quality, garp, growth, quant, contrarian, activist
	KindMomentum    Kind = "momentum"    // trend_momentum, macro_momentum
	KindAllocation  Kind = "allocation"  // macro_risk_parity
)

// KindFor maps a declared style to its evaluator kind. Unknown styles default to
// fundamental (the most common family).
func KindFor(style string) Kind {
	switch style {
	case "macro_risk_parity":
		return KindAllocation
	case "trend_momentum", "macro_momentum":
		return KindMomentum
	default:
		return KindFundamental
	}
}

// Kind returns the evaluator kind for this strategy.
func (c Config) Kind() Kind { return KindFor(c.Identity.Style) }

// Load reads and validates a single strategy file.
func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("strategy: reading %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("strategy: parsing %s: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("strategy: invalid %s: %w", path, err)
	}
	return cfg, nil
}

// LoadDir loads every v2 strategy file in dir, sorted by id. Files that are not
// schema v2 (e.g. the legacy benjamin.yml / loose.yml fixtures) are skipped, so
// the directory can hold both during the transition.
func LoadDir(dir string) ([]Config, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("strategy: globbing %s: %w", dir, err)
	}
	var configs []Config
	for _, path := range matches {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("strategy: reading %s: %w", path, err)
		}
		var cfg Config
		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("strategy: parsing %s: %w", path, err)
		}
		if cfg.SchemaVersion != 2 {
			continue // skip legacy/non-v2 files
		}
		if err := cfg.validate(); err != nil {
			return nil, fmt.Errorf("strategy: invalid %s: %w", path, err)
		}
		configs = append(configs, cfg)
	}
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Identity.ID < configs[j].Identity.ID
	})
	return configs, nil
}

func (c Config) validate() error {
	if c.SchemaVersion != 2 {
		return fmt.Errorf("unsupported schema_version %d (want 2)", c.SchemaVersion)
	}
	if c.Identity.ID == "" {
		return fmt.Errorf("identity.id is required")
	}
	if c.Identity.Style == "" {
		return fmt.Errorf("identity.style is required")
	}
	if c.Risk.MaxPositionSize <= 0 || c.Risk.MaxPositionSize > 1 {
		return fmt.Errorf("risk.max_position_size must be in (0,1], got %v", c.Risk.MaxPositionSize)
	}
	return nil
}
