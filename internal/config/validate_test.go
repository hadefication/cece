package config

import "testing"

func TestValidateName_Valid(t *testing.T) {
	for _, name := range []string{"work", "my-profile", "client_1", "A", "test123"} {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", name, err)
		}
	}
}

func TestValidateName_Invalid(t *testing.T) {
	for _, name := range []string{"", "has space", "has;semicolon", "has'quote", "../traversal", "has/slash", "-starts-with-dash"} {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", name)
		}
	}
}

func TestValidatePlugin_Valid(t *testing.T) {
	for _, p := range []string{"plugin:imessage@claude-plugins-official", "my-plugin", "test_plugin/v2"} {
		if err := ValidatePlugin(p); err != nil {
			t.Errorf("ValidatePlugin(%q) = %v, want nil", p, err)
		}
	}
}

func TestValidatePlugin_Invalid(t *testing.T) {
	for _, p := range []string{"", "has space", "has;semicolon", "has'quote", "$(inject)", "`inject`"} {
		if err := ValidatePlugin(p); err == nil {
			t.Errorf("ValidatePlugin(%q) = nil, want error", p)
		}
	}
}
