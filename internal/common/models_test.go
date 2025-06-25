package common

import "testing"

func TestStringPtr(t *testing.T) {
	str := "test"
	ptr := StringPtr(str)
	if ptr == nil || *ptr != str {
		t.Errorf("StringPtr(%q) = %v, want pointer to %q", str, ptr, str)
	}
}

func TestIntPtr(t *testing.T) {
	val := 42
	ptr := IntPtr(val)
	if ptr == nil || *ptr != val {
		t.Errorf("IntPtr(%d) = %v, want pointer to %d", val, ptr, val)
	}
}

func TestBoolPtr(t *testing.T) {
	val := true
	ptr := BoolPtr(val)
	if ptr == nil || *ptr != val {
		t.Errorf("BoolPtr(%v) = %v, want pointer to %v", val, ptr, val)
	}
}
