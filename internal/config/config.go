package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// IPPair represents a source and destination IP address pair.
type IPPair struct {
	SrcIP string `yaml:"src_ip"`
	DstIP string `yaml:"dst_ip"`
}

// Validate checks that both IPs are present and valid.
func (p *IPPair) Validate() error {
	if p.SrcIP == "" {
		return fmt.Errorf("source IP is required")
	}
	if p.DstIP == "" {
		return fmt.Errorf("destination IP is required")
	}
	if net.ParseIP(p.SrcIP) == nil {
		return fmt.Errorf("invalid source IP: %s", p.SrcIP)
	}
	if net.ParseIP(p.DstIP) == nil {
		return fmt.Errorf("invalid destination IP: %s", p.DstIP)
	}
	return nil
}

// Target represents a single send target with outer and inner IP pairs.
type Target struct {
	RealIP  IPPair `yaml:"real_ip"`
	ErrorIP IPPair `yaml:"error_ip"`
}

// Validate checks that both IP pairs are valid.
func (t *Target) Validate() error {
	if err := t.RealIP.Validate(); err != nil {
		return fmt.Errorf("real_ip: %w", err)
	}
	if err := t.ErrorIP.Validate(); err != nil {
		return fmt.Errorf("error_ip: %w", err)
	}
	return nil
}

// Profile is a named group of targets.
type Profile struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Targets     []Target `yaml:"targets"`
}

// Validate checks that the profile has a name and at least one valid target.
func (p *Profile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("profile name is required")
	}
	if len(p.Targets) == 0 {
		return fmt.Errorf("profile %q must have at least one target", p.Name)
	}
	for i := range p.Targets {
		if err := p.Targets[i].Validate(); err != nil {
			return fmt.Errorf("profile %q target[%d]: %w", p.Name, i, err)
		}
	}
	return nil
}

// Options holds global defaults that can be overridden by CLI flags.
type Options struct {
	TTL     uint8 `yaml:"ttl"`
	Verbose bool  `yaml:"verbose"`
	DryRun  bool  `yaml:"dry_run"`
}

// Config is the top-level configuration structure.
type Config struct {
	Profiles []Profile `yaml:"profiles"`
	Options  Options   `yaml:"options"`
}

// Validate checks the entire configuration for correctness.
func (c *Config) Validate() error {
	if len(c.Profiles) == 0 {
		return fmt.Errorf("at least one profile is required")
	}

	seen := make(map[string]bool, len(c.Profiles))
	for i := range c.Profiles {
		if err := c.Profiles[i].Validate(); err != nil {
			return fmt.Errorf("profiles[%d]: %w", i, err)
		}
		if seen[c.Profiles[i].Name] {
			return fmt.Errorf("duplicate profile name: %q", c.Profiles[i].Name)
		}
		seen[c.Profiles[i].Name] = true
	}
	return nil
}

// GetProfile returns the profile with the given name.
// If name is empty, it returns the first profile named "default",
// or the first profile if no default exists.
func (c *Config) GetProfile(name string) (*Profile, error) {
	if name == "" {
		name = "default"
	}
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i], nil
		}
	}
	return nil, fmt.Errorf("profile %q not found", name)
}

// DefaultConfigDir returns the default directory for icer config files.
// Respects $XDG_CONFIG_HOME, falls back to ~/.config/icer.
func DefaultConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "icer")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "icer")
}

// DefaultConfigPath returns the full path to the default config file.
func DefaultConfigPath() string {
	dir := DefaultConfigDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "config.yaml")
}

// ResolveConfigPath determines which config file to use.
// Priority: explicit path > $ICER_CONFIG > XDG default.
// Returns empty string if no config file is found (which is OK for CLI-only usage).
func ResolveConfigPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if envPath := os.Getenv("ICER_CONFIG"); envPath != "" {
		return envPath
	}
	defaultPath := DefaultConfigPath()
	if defaultPath != "" {
		if _, err := os.Stat(defaultPath); err == nil {
			return defaultPath
		}
	}
	return ""
}

// Load reads and parses a config file. Returns nil config (not an error)
// if path is empty, meaning no config file is available.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	if cfg.Options.TTL == 0 {
		cfg.Options.TTL = 64
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return &cfg, nil
}

// SampleConfig returns a sample YAML configuration string.
const SampleConfig = `# icer configuration file
# Place at ~/.config/icer/config.yaml or set $ICER_CONFIG

profiles:
  - name: "default"
    description: "Default testing profile"
    targets:
      - real_ip:
          src_ip: "192.168.10.10"
          dst_ip: "192.168.20.20"
        error_ip:
          src_ip: "10.10.10.10"
          dst_ip: "20.20.20.20"

  - name: "web-servers"
    description: "Test against multiple web servers"
    targets:
      - real_ip: { src_ip: "192.168.1.1", dst_ip: "192.168.1.10" }
        error_ip: { src_ip: "10.0.0.1", dst_ip: "10.0.0.10" }
      - real_ip: { src_ip: "192.168.1.1", dst_ip: "192.168.1.11" }
        error_ip: { src_ip: "10.0.0.1", dst_ip: "10.0.0.11" }

options:
  ttl: 64
  verbose: false
  dry_run: false
`
