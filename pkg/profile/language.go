package profile

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/pprof/profile"
)

type language string

const (
	langGo     language = "Go"
	langCPP    language = "C++"
	langRust   language = "Rust"
	langJava   language = "Java"
	langC      language = "C"
	langUnknown language = "Unknown"
)

// languagePatterns maps languages to patterns matched against function names.
var languagePatterns = map[language][]*regexp.Regexp{
	langGo: {
		regexp.MustCompile(`^runtime\.`),
		regexp.MustCompile(`^(main|fmt|net/http|encoding|io|os|sync|context)\.`),
		regexp.MustCompile(`\bgo\b`),
	},
	langCPP: {
		regexp.MustCompile(`^std::`),
		regexp.MustCompile(`::`),
		regexp.MustCompile(`^llvm::`),
		regexp.MustCompile(`^boost::`),
	},
	langRust: {
		regexp.MustCompile(`^core::`),
		regexp.MustCompile(`^alloc::`),
		regexp.MustCompile(`^std::`),
		regexp.MustCompile(`^<.*as .+>::`),
		regexp.MustCompile(`\bcrate::`),
	},
	langJava: {
		regexp.MustCompile(`^java\.lang\.`),
		regexp.MustCompile(`^java\.util\.`),
		regexp.MustCompile(`^(com|org|net)\.`),
		regexp.MustCompile(`^javax?\.`),
	},
	langC: {
		regexp.MustCompile(`^[a-z_]\w+$`), // simple C-style function names
	},
}

// filenamePatterns maps languages to file extension patterns.
var filenamePatterns = map[language][]string{
	langGo:   {".go"},
	langCPP:  {".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".h"},
	langRust: {".rs"},
	langJava: {".java"},
	langC:    {".c"},
}

// DetectLanguage analyzes function names and filenames in the profile to detect
// the programming language. Returns the language with the most matches.
func DetectLanguage(p *profile.Profile) string {
	if p == nil || len(p.Function) == 0 {
		return string(langUnknown)
	}

	scores := map[language]int{}

	for _, fn := range p.Function {
		scoreFunction(fn.Name, scores)
		scoreFilename(fn.Filename, scores)
	}

	return bestMatch(scores)
}

func scoreFunction(name string, scores map[language]int) {
	for lang, patterns := range languagePatterns {
		for _, re := range patterns {
			if re.MatchString(name) {
				scores[lang]++
			}
		}
	}
}

func scoreFilename(filename string, scores map[language]int) {
	if filename == "" {
		return
	}
	ext := strings.ToLower(filepath.Ext(filename))
	for lang, extensions := range filenamePatterns {
		for _, e := range extensions {
			if ext == e {
				// C and C++ both match .h; C++ patterns are more specific,
				// so only count .h for C if C++ isn't already scoring higher.
				if lang == langC && ext == ".h" && scores[langCPP] > 0 {
					continue
				}
				scores[lang]++
			}
		}
	}
}

func bestMatch(scores map[language]int) string {
	best := langUnknown
	bestScore := 0

	for lang, score := range scores {
		if score > bestScore {
			bestScore = score
			best = lang
		}
	}

	// Require at least one match; otherwise unknown
	if bestScore == 0 {
		return string(langUnknown)
	}
	return string(best)
}
