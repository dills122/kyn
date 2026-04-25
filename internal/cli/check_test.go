package cli

import "testing"

func TestValidateCheckOptions(t *testing.T) {
	valid := checkOptions{
		Format: "text",
		FailOn: "error",
	}

	tests := []struct {
		name    string
		modify  func(*checkOptions)
		wantErr bool
	}{
		{
			name: "valid files mode",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
			},
		},
		{
			name: "valid files-from mode",
			modify: func(o *checkOptions) {
				o.FilesFrom = "changed.txt"
			},
		},
		{
			name: "valid git mode",
			modify: func(o *checkOptions) {
				o.Base = "origin/main"
				o.Head = "HEAD"
			},
		},
		{
			name:    "invalid no mode",
			wantErr: true,
		},
		{
			name: "invalid multiple modes",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.FilesFrom = "changed.txt"
			},
			wantErr: true,
		},
		{
			name: "invalid partial git mode",
			modify: func(o *checkOptions) {
				o.Base = "origin/main"
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "xml"
			},
			wantErr: true,
		},
		{
			name: "invalid fail-on",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.FailOn = "info"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := valid
			if tt.modify != nil {
				tt.modify(&o)
			}

			err := validateCheckOptions(o)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
