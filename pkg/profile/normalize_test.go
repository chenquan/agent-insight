package profile

import "testing"

func TestNormalizeMappingFile(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/binary", "binary"},
		{"/home/rsilvera/cppbench/cppbench_server_main", "cppbench_server_main"},
		{"cppbench_server_main", "cppbench_server_main"},
		{"/lib/libc-2.15.so", "libc-2.15.so"},
		{"[vdso]", "[vdso]"},
		{"[vsyscall]", "[vsyscall]"},
		{"", ""},
	}
	for _, tt := range tests {
		got := normalizeMappingFile(tt.input)
		if got != tt.want {
			t.Errorf("normalizeMappingFile(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
