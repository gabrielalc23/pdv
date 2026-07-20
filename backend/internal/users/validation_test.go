package users

import "testing"

func TestNormalizeDisplayName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid and trimmed", input: "  Novo Nome  ", want: "Novo Nome"},
		{name: "blank", input: " \t ", wantErr: true},
		{name: "control", input: "Nome\u0000", wantErr: true},
		{name: "too long", input: string(make([]rune, 151)), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeDisplayName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeDisplayName() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}
