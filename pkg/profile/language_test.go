package profile

import (
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestDetectLanguage_Go(t *testing.T) {
	p := &pprofprofile.Profile{
		Function: []*pprofprofile.Function{
			{Name: "runtime.mallocgc", Filename: "runtime/malloc.go"},
			{Name: "main.handleRequest", Filename: "main.go"},
			{Name: "encoding/json.Marshal", Filename: "encoding/json/encode.go"},
		},
	}
	got := DetectLanguage(NewProfile(p))
	if got != string(langGo) {
		t.Errorf("DetectLanguage() = %q, want %q", got, langGo)
	}
}

func TestDetectLanguage_CPP(t *testing.T) {
	p := &pprofprofile.Profile{
		Function: []*pprofprofile.Function{
			{Name: "std::vector<int>::push_back", Filename: "stl/vector.cpp"},
			{Name: "MyApp::process", Filename: "src/app.cpp"},
			{Name: "llvm::IRBuilder", Filename: "ir/irbuilder.cpp"},
		},
	}
	got := DetectLanguage(NewProfile(p))
	if got != string(langCPP) {
		t.Errorf("DetectLanguage() = %q, want %q", got, langCPP)
	}
}

func TestDetectLanguage_Rust(t *testing.T) {
	p := &pprofprofile.Profile{
		Function: []*pprofprofile.Function{
			{Name: "core::alloc::alloc", Filename: "core/alloc/mod.rs"},
			{Name: "mycrate::process", Filename: "src/main.rs"},
			{Name: "<T as std::fmt::Display>::fmt", Filename: "lib.rs"},
		},
	}
	got := DetectLanguage(NewProfile(p))
	if got != string(langRust) {
		t.Errorf("DetectLanguage() = %q, want %q", got, langRust)
	}
}

func TestDetectLanguage_Java(t *testing.T) {
	p := &pprofprofile.Profile{
		Function: []*pprofprofile.Function{
			{Name: "java.lang.String.getBytes", Filename: "String.java"},
			{Name: "com.example.Service.process", Filename: "Service.java"},
			{Name: "org.apache.http.client.execute", Filename: "HttpClient.java"},
		},
	}
	got := DetectLanguage(NewProfile(p))
	if got != string(langJava) {
		t.Errorf("DetectLanguage() = %q, want %q", got, langJava)
	}
}

func TestDetectLanguage_C(t *testing.T) {
	p := &pprofprofile.Profile{
		Function: []*pprofprofile.Function{
			{Name: "malloc", Filename: "malloc.c"},
			{Name: "memcpy", Filename: "string.c"},
			{Name: "handle_request", Filename: "main.c"},
		},
	}
	got := DetectLanguage(NewProfile(p))
	if got != string(langC) {
		t.Errorf("DetectLanguage() = %q, want %q", got, langC)
	}
}

func TestDetectLanguage_Unknown(t *testing.T) {
	tests := []struct {
		name string
		p    *pprofprofile.Profile
	}{
		{"nil profile", nil},
		{"empty profile", &pprofprofile.Profile{}},
		{"no symbols", &pprofprofile.Profile{
			Function: []*pprofprofile.Function{},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLanguage(NewProfile(tt.p))
			if got != string(langUnknown) {
				t.Errorf("DetectLanguage() = %q, want %q", got, langUnknown)
			}
		})
	}
}

func TestDetectLanguage_Mixed_GoWins(t *testing.T) {
	p := &pprofprofile.Profile{
		Function: []*pprofprofile.Function{
			{Name: "runtime.mallocgc", Filename: "runtime/malloc.go"},
			{Name: "main.main", Filename: "main.go"},
			{Name: "encoding/json.Marshal", Filename: "encoding/json/encode.go"},
		},
	}
	got := DetectLanguage(NewProfile(p))
	if got != string(langGo) {
		t.Errorf("DetectLanguage() = %q, want %q (Go should win with more matches)", got, langGo)
	}
}
