package main

import "testing"

func TestRunHelp(t *testing.T) {
	if code := run([]string{"--help"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if code := run([]string{"crawl", "--help"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if code := run([]string{"unknown"}); code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}

func TestRunMissingRequiredFlags(t *testing.T) {
	if code := run([]string{"crawl"}); code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}
