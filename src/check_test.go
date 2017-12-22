package main

import "testing"
import "runtime"

func TestSliceContainString(t *testing.T) {
	slc := []string{"hello", "world"}
	s := "world"
	pc, file, line, ok := runtime.Caller(0)
	if !sliceContainString(slc, s) {
		t.Errorf("failed to pass unit test,function %s,file %s,line %d ok %v", runtime.FuncForPC(pc).Name(), file, line, ok)
	}
}
