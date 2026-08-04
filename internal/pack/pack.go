package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Primitive   string `yaml:"primitive"`
	Kind        string `yaml:"kind"`
	Scoped      bool   `yaml:"scoped"`
	Description string `yaml:"description"`
}

type RankPolicy struct {
	PrimitiveWeights map[string]int     `yaml:"primitive_weights"`
	ScopeBoost       map[string]float64 `yaml:"scope_boost"`
}

type EntityDef struct {
	Name      string `yaml:"name"`
	PrimaryID string `yaml:"primary_id"`
}

type StateTransition struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type StateMachine struct {
	States      []string          `yaml:"states"`
	Transitions []StateTransition `yaml:"transitions"`
}

type Pack struct {
	ID              string                       `yaml:"id"`
	Version         string                       `yaml:"version"`
	Description     string                       `yaml:"description"`
	Vocabulary      map[string]Entry             `yaml:"vocabulary"`
	MetadataSchemas map[string]map[string]any    `yaml:"metadata_schemas"`
	LifecycleRules  []LifecycleRule              `yaml:"lifecycle_rules"`
	RankPolicy      RankPolicy                   `yaml:"rank_policy"`
	EvalFixtures    string                       `yaml:"eval_fixtures"`
	Entities        []EntityDef                  `yaml:"entities,omitempty"`
	StateMachines   map[string]StateMachine      `yaml:"state_machines,omitempty"`
}

type Registry struct {
	byID map[string]*Pack
}

func NewRegistry() *Registry {
	return &Registry{byID: map[string]*Pack{}}
}

func (r *Registry) Register(p *Pack) {
	if p == nil || p.ID == "" {
		return
	}
	if existing, ok := r.byID[p.ID]; ok && versionLess(p.Version, existing.Version) {
		return
	}
	r.byID[p.ID] = p
}

func versionLess(a, b string) bool {
	ai, aErr := strconv.Atoi(strings.TrimSpace(a))
	bi, bErr := strconv.Atoi(strings.TrimSpace(b))
	if aErr == nil && bErr == nil {
		return ai < bi
	}
	return strings.TrimSpace(a) < strings.TrimSpace(b)
}

func (r *Registry) Get(id string) (*Pack, bool) {
	if id == "" || id == "core" {
		return nil, false
	}
	p, ok := r.byID[id]
	return p, ok
}

func LoadFile(path string) (*Pack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Pack
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse pack %s: %w", path, err)
	}
	if p.ID == "" {
		return nil, fmt.Errorf("pack %s: missing id", path)
	}
	if err := loadPackSidecars(filepath.Dir(path), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func loadPackSidecars(dir string, p *Pack) error {
	entitiesPath := filepath.Join(dir, "entities.yaml")
	if raw, err := os.ReadFile(entitiesPath); err == nil {
		var wrap struct {
			Entities []EntityDef `yaml:"entities"`
		}
		if err := yaml.Unmarshal(raw, &wrap); err != nil {
			return fmt.Errorf("parse %s: %w", entitiesPath, err)
		}
		if len(wrap.Entities) > 0 {
			p.Entities = wrap.Entities
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	smPath := filepath.Join(dir, "state-machines.yaml")
	if raw, err := os.ReadFile(smPath); err == nil {
		var machines map[string]StateMachine
		if err := yaml.Unmarshal(raw, &machines); err != nil {
			return fmt.Errorf("parse %s: %w", smPath, err)
		}
		if len(machines) > 0 {
			if p.StateMachines == nil {
				p.StateMachines = map[string]StateMachine{}
			}
			for name, sm := range machines {
				p.StateMachines[name] = sm
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func LoadRegistryFromDir(root string) (*Registry, error) {
	reg := NewRegistry()
	if root == "" {
		return reg, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("packs root is not a directory: %s", root)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, vertical := range entries {
		if !vertical.IsDir() {
			continue
		}
		versions, err := os.ReadDir(filepath.Join(root, vertical.Name()))
		if err != nil {
			return nil, err
		}
		for _, version := range versions {
			if !version.IsDir() {
				continue
			}
			path := filepath.Join(root, vertical.Name(), version.Name(), "pack.yaml")
			if _, err := os.Stat(path); err != nil {
				continue
			}
			p, err := LoadFile(path)
			if err != nil {
				return nil, err
			}
			reg.Register(p)
		}
	}
	return reg, nil
}

// LabelForKind returns the default pack label and primitive for a legacy kind.
func (p *Pack) LabelForKind(kind string) (label, primitive string, ok bool) {
	if p == nil {
		return "", "", false
	}
	preferred := map[string]string{
		"preference": "voice_profile",
		"profile":    "profile",
		"fact":       "fact",
	}
	if preferredLabel, exists := preferred[kind]; exists {
		if entry, ok := p.Vocabulary[preferredLabel]; ok {
			return preferredLabel, entry.Primitive, true
		}
	}
	for label, entry := range p.Vocabulary {
		if entry.Kind == kind {
			return label, entry.Primitive, true
		}
	}
	return "", "", false
}

func (p *Pack) PrimitiveWeight(primitive string) float64 {
	if p == nil || primitive == "" {
		return 0
	}
	w, ok := p.RankPolicy.PrimitiveWeights[primitive]
	if !ok || w <= 0 {
		return 0
	}
	return float64(w) / 100.0
}

func NormalizePrimitive(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ValidateStateTransition checks a named FSM transition.
// Empty from allows bootstrap into any known state.
func (p *Pack) ValidateStateTransition(machine, from, to string) error {
	if p == nil || machine == "" || to == "" {
		return nil
	}
	sm, ok := p.StateMachines[machine]
	if !ok {
		return nil
	}
	to = strings.ToLower(strings.TrimSpace(to))
	from = strings.ToLower(strings.TrimSpace(from))
	if !stateContains(sm.States, to) {
		return fmt.Errorf("state machine %q: unknown state %q", machine, to)
	}
	if from == "" {
		return nil
	}
	if !stateContains(sm.States, from) {
		return fmt.Errorf("state machine %q: unknown from state %q", machine, from)
	}
	for _, tr := range sm.Transitions {
		if strings.EqualFold(tr.From, from) && strings.EqualFold(tr.To, to) {
			return nil
		}
	}
	return fmt.Errorf("state machine %q: transition %s → %s not allowed", machine, from, to)
}

func stateContains(states []string, want string) bool {
	for _, s := range states {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}

// MachineForLabel maps pack labels onto sidecar FSM names.
func (p *Pack) MachineForLabel(label string) string {
	switch strings.TrimSpace(label) {
	case "ticket_state":
		if _, ok := p.StateMachines["ticket_status"]; ok {
			return "ticket_status"
		}
	}
	return ""
}
