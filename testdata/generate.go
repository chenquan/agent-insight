package main

import (
	"os"

	"github.com/google/pprof/profile"
)

func main() {
	generateCPUProfile("testdata/cpu.pb.gz")
	generateHeapProfile("testdata/heap.pb.gz")
	generateGoroutineProfile("testdata/goroutine.pb.gz")
}

func generateCPUProfile(path string) {
	p := &profile.Profile{
		PeriodType: &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:     10000000, // 10ms
		DurationNanos: 30e9,  // 30s
		SampleType: []*profile.ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
	}

	// Functions
	mallocgc := &profile.Function{ID: 1, Name: "runtime.mallocgc", SystemName: "runtime.mallocgc", Filename: "runtime/malloc.go"}
	mainFunc := &profile.Function{ID: 2, Name: "main.main", SystemName: "main.main", Filename: "main.go"}
	handleReq := &profile.Function{ID: 3, Name: "main.handleRequest", SystemName: "main.handleRequest", Filename: "main.go"}
	jsonMarshal := &profile.Function{ID: 4, Name: "encoding/json.Marshal", SystemName: "encoding/json.Marshal", Filename: "encoding/json/encode.go"}
	readAll := &profile.Function{ID: 5, Name: "io.ReadAll", SystemName: "io.ReadAll", Filename: "io/io.go"}

	p.Function = []*profile.Function{mallocgc, mainFunc, handleReq, jsonMarshal, readAll}

	// Mappings
	mapping := &profile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}
	p.Mapping = []*profile.Mapping{mapping}

	// Locations
	locMallocgc := &profile.Location{ID: 1, Mapping: mapping, Address: 0x1100, Line: []profile.Line{{Function: mallocgc, Line: 1020}}}
	locMain := &profile.Location{ID: 2, Mapping: mapping, Address: 0x1200, Line: []profile.Line{{Function: mainFunc, Line: 15}}}
	locHandleReq := &profile.Location{ID: 3, Mapping: mapping, Address: 0x1300, Line: []profile.Line{{Function: handleReq, Line: 42}}}
	locJsonMarshal := &profile.Location{ID: 4, Mapping: mapping, Address: 0x1400, Line: []profile.Line{{Function: jsonMarshal, Line: 160}}}
	locReadAll := &profile.Location{ID: 5, Mapping: mapping, Address: 0x1500, Line: []profile.Line{{Function: readAll, Line: 88}}}
	locNoSym := &profile.Location{ID: 6, Mapping: mapping, Address: 0x1600} // No symbol info

	p.Location = []*profile.Location{locMallocgc, locMain, locHandleReq, locJsonMarshal, locReadAll, locNoSym}

	// Samples (leaf first)
	p.Sample = []*profile.Sample{
		// mallocgc called from handleRequest called from main
		{Location: []*profile.Location{locMallocgc, locHandleReq, locMain}, Value: []int64{500, 5000000000}},
		// json.Marshal called from handleRequest called from main
		{Location: []*profile.Location{locJsonMarshal, locHandleReq, locMain}, Value: []int64{300, 3000000000}},
		// ReadAll called from handleRequest called from main
		{Location: []*profile.Location{locReadAll, locHandleReq, locMain}, Value: []int64{150, 1500000000}},
		// mallocgc directly from main
		{Location: []*profile.Location{locMallocgc, locMain}, Value: []int64{100, 1000000000}},
		// No-symbol location
		{Location: []*profile.Location{locNoSym, locMain}, Value: []int64{50, 500000000}},
	}

	writeProfile(path, p)
}

func generateHeapProfile(path string) {
	p := &profile.Profile{
		PeriodType: &profile.ValueType{Type: "space", Unit: "bytes"},
		Period:     524288, // 512KB
		SampleType: []*profile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
			{Type: "inuse_objects", Unit: "count"},
			{Type: "inuse_space", Unit: "bytes"},
		},
	}

	mallocgc := &profile.Function{ID: 1, Name: "runtime.mallocgc", SystemName: "runtime.mallocgc", Filename: "runtime/malloc.go"}
	mainFunc := &profile.Function{ID: 2, Name: "main.main", SystemName: "main.main", Filename: "main.go"}
	makeSlice := &profile.Function{ID: 3, Name: "main.makeSlice", SystemName: "main.makeSlice", Filename: "main.go"}
	newBuf := &profile.Function{ID: 4, Name: "bytes.NewBuffer", SystemName: "bytes.NewBuffer", Filename: "bytes/buffer.go"}

	p.Function = []*profile.Function{mallocgc, mainFunc, makeSlice, newBuf}

	mapping := &profile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}
	p.Mapping = []*profile.Mapping{mapping}

	locMallocgc := &profile.Location{ID: 1, Mapping: mapping, Address: 0x1100, Line: []profile.Line{{Function: mallocgc, Line: 1020}}}
	locMain := &profile.Location{ID: 2, Mapping: mapping, Address: 0x1200, Line: []profile.Line{{Function: mainFunc, Line: 15}}}
	locMakeSlice := &profile.Location{ID: 3, Mapping: mapping, Address: 0x1300, Line: []profile.Line{{Function: makeSlice, Line: 30}}}
	locNewBuf := &profile.Location{ID: 4, Mapping: mapping, Address: 0x1400, Line: []profile.Line{{Function: newBuf, Line: 50}}}

	p.Location = []*profile.Location{locMallocgc, locMain, locMakeSlice, locNewBuf}

	p.Sample = []*profile.Sample{
		// makeSlice allocates a lot
		{Location: []*profile.Location{locMakeSlice, locMain}, Value: []int64{1000, 838860800, 500, 419430400}},
		// bytes.NewBuffer
		{Location: []*profile.Location{locNewBuf, locMain}, Value: []int64{500, 8388608, 200, 4194304}},
		// direct mallocgc
		{Location: []*profile.Location{locMallocgc, locMain}, Value: []int64{200, 1048576, 100, 524288}},
	}

	writeProfile(path, p)
}

func writeProfile(path string, p *profile.Profile) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := p.Write(f); err != nil {
		panic(err)
	}
}

// generateGoroutineProfile builds a goroutine profile carrying pprof labels
// (state / wait_reason) on each sample, mirroring what the Go runtime emits.
// It is the canonical test fixture for the tags command and --tag filtering.
func generateGoroutineProfile(path string) {
	p := &profile.Profile{
		PeriodType:    &profile.ValueType{Type: "goroutine", Unit: "count"},
		Period:        1,
		DurationNanos: 0, // goroutine profiles are instantaneous
		SampleType:    []*profile.ValueType{{Type: "goroutine", Unit: "count"}},
	}

	// Functions
	queryCtx := &profile.Function{ID: 1, Name: "database/sql.(*DB).QueryContext", SystemName: "database/sql.(*DB).QueryContext", Filename: "database/sql/sql.go"}
	handleReq := &profile.Function{ID: 2, Name: "main.handleRequest", SystemName: "main.handleRequest", Filename: "main.go"}
	mainFunc := &profile.Function{ID: 3, Name: "main.main", SystemName: "main.main", Filename: "main.go"}
	signote := &profile.Function{ID: 4, Name: "runtime.sigNoteSleep", SystemName: "runtime.sigNoteSleep", Filename: "runtime/signal_unix.go"}
	selectgo := &profile.Function{ID: 5, Name: "runtime.selectgo", SystemName: "runtime.selectgo", Filename: "runtime/select.go"}
	mallocgc := &profile.Function{ID: 6, Name: "runtime.mallocgc", SystemName: "runtime.mallocgc", Filename: "runtime/malloc.go"}
	connRead := &profile.Function{ID: 7, Name: "net/http.(*connReader).Read", SystemName: "net/http.(*connReader).Read", Filename: "net/http/server.go"}

	p.Function = []*profile.Function{queryCtx, handleReq, mainFunc, signote, selectgo, mallocgc, connRead}

	mapping := &profile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}
	p.Mapping = []*profile.Mapping{mapping}

	locQuery := &profile.Location{ID: 1, Mapping: mapping, Address: 0x1100, Line: []profile.Line{{Function: queryCtx, Line: 2200}}}
	locHandleReq := &profile.Location{ID: 2, Mapping: mapping, Address: 0x1200, Line: []profile.Line{{Function: handleReq, Line: 42}}}
	locMain := &profile.Location{ID: 3, Mapping: mapping, Address: 0x1300, Line: []profile.Line{{Function: mainFunc, Line: 15}}}
	locSignote := &profile.Location{ID: 4, Mapping: mapping, Address: 0x1400, Line: []profile.Line{{Function: signote, Line: 100}}}
	locSelectgo := &profile.Location{ID: 5, Mapping: mapping, Address: 0x1500, Line: []profile.Line{{Function: selectgo, Line: 300}}}
	locMallocgc := &profile.Location{ID: 6, Mapping: mapping, Address: 0x1600, Line: []profile.Line{{Function: mallocgc, Line: 1020}}}
	locRead := &profile.Location{ID: 7, Mapping: mapping, Address: 0x1700, Line: []profile.Line{{Function: connRead, Line: 800}}}

	p.Location = []*profile.Location{locQuery, locHandleReq, locMain, locSignote, locSelectgo, locMallocgc, locRead}

	// Samples (leaf first). Each carries a goroutine runtime label set.
	// Label semantics: state in {blocked, running, syscall}; wait_reason in
	// {IO, semacquire, channel} (only meaningful when state=blocked).
	p.Sample = []*profile.Sample{
		{Location: []*profile.Location{locQuery, locHandleReq, locMain}, Label: map[string][]string{"state": {"blocked"}, "wait_reason": {"IO"}}, Value: []int64{1200}},
		{Location: []*profile.Location{locQuery, locHandleReq, locMain}, Label: map[string][]string{"state": {"blocked"}, "wait_reason": {"semacquire"}}, Value: []int64{800}},
		{Location: []*profile.Location{locQuery, locHandleReq, locMain}, Label: map[string][]string{"state": {"running"}}, Value: []int64{1500}},
		{Location: []*profile.Location{locQuery, locHandleReq, locMain}, Label: map[string][]string{"state": {"syscall"}}, Value: []int64{600}},
		{Location: []*profile.Location{locSignote, locMain}, Label: map[string][]string{"state": {"syscall"}}, Value: []int64{400}},
		{Location: []*profile.Location{locSelectgo, locHandleReq, locMain}, Label: map[string][]string{"state": {"blocked"}, "wait_reason": {"channel"}}, Value: []int64{1000}},
		{Location: []*profile.Location{locMallocgc, locHandleReq, locMain}, Label: map[string][]string{"state": {"running"}}, Value: []int64{200}},
		{Location: []*profile.Location{locRead, locHandleReq, locMain}, Label: map[string][]string{"state": {"blocked"}, "wait_reason": {"IO"}}, Value: []int64{500}},
	}

	writeProfile(path, p)
}
