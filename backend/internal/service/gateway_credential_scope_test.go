package service

import "testing"

func TestIsClaudeCodeCredentialScopeError(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{
			name: "original scope error",
			msg:  "This key is only authorized for use with claude code and cannot be used for other api requests",
			want: true,
		},
		{
			name: "third-party extra usage error",
			msg:  "Third-party apps now draw from your extra usage, not your plan limits. Add more at claude.ai/settings/usage and keep going.",
			want: true,
		},
		{
			name: "case insensitive",
			msg:  "THIRD-PARTY APPS NOW DRAW FROM YOUR EXTRA USAGE, NOT YOUR PLAN LIMITS.",
			want: true,
		},
		{
			name: "unrelated extra usage message",
			msg:  "Extra usage is required for long context requests.",
			want: false,
		},
		{
			name: "empty string",
			msg:  "",
			want: false,
		},
		{
			name: "unrelated error",
			msg:  "Invalid API key provided.",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaudeCodeCredentialScopeError(tt.msg)
			if got != tt.want {
				t.Errorf("isClaudeCodeCredentialScopeError(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}
