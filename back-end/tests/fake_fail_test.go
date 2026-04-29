package tests

import "testing"

func TestDeliberateFailure(t *testing.T) {
	t.Fatal("this is a deliberate failure to verify CI catches broken tests")
}
