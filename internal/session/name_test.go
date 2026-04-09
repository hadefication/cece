package session

import (
	"regexp"
	"testing"
)

func TestGenerateName_HomeDir_NoProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "", "/Users/inggo", "/Users/inggo")
	pattern := `^inggo@Mac-mini-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestGenerateName_ProjectDir_NoProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "", "/Users/inggo/Code/myproject", "/Users/inggo")
	pattern := `^inggo@Mac-mini-myproject-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestGenerateName_HomeDir_WithProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "work", "/Users/inggo", "/Users/inggo")
	pattern := `^inggo@Mac-mini-work-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestGenerateName_ProjectDir_WithProfile(t *testing.T) {
	name := GenerateName("inggo", "Mac-mini", "work", "/Users/inggo/Code/myproject", "/Users/inggo")
	pattern := `^inggo@Mac-mini-work-myproject-[a-z]{3}\d{2}\d{4}-\d{4}$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name %q does not match pattern %q", name, pattern)
	}
}

func TestTmuxRemoteName_NoProfile(t *testing.T) {
	got := TmuxRemoteName("", "myproject")
	if got != "cece-remote-myproject" {
		t.Errorf("got %q, want %q", got, "cece-remote-myproject")
	}
}

func TestTmuxRemoteName_WithProfile(t *testing.T) {
	got := TmuxRemoteName("work", "myproject")
	if got != "cece-remote-work-myproject" {
		t.Errorf("got %q, want %q", got, "cece-remote-work-myproject")
	}
}

func TestTmuxChannelName_NoProfile(t *testing.T) {
	got := TmuxChannelName("", "imessage")
	if got != "cece-channel-imessage" {
		t.Errorf("got %q, want %q", got, "cece-channel-imessage")
	}
}

func TestTmuxChannelName_WithProfile(t *testing.T) {
	got := TmuxChannelName("work", "imessage")
	if got != "cece-channel-work-imessage" {
		t.Errorf("got %q, want %q", got, "cece-channel-work-imessage")
	}
}

func TestDetectMachine(t *testing.T) {
	machine := DetectMachine()
	if machine == "" {
		t.Error("DetectMachine() returned empty string")
	}
}
