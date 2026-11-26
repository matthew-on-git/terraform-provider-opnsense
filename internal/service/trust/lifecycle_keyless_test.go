// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import "testing"

func TestTrustPackageVersionAtLeast(t *testing.T) {
	t.Parallel()
	rows := []struct {
		actual, minimum string
		want            bool
		fail            bool
	}{
		{"26.7.8", "26.7.8_2", false, false},
		{"26.7.8_1", "26.7.8_2", false, false},
		{"26.7.8_2", "26.7.8_2", true, false},
		{"26.7.8_3", "26.7.8_2", true, false},
		{"26.7.9", "26.7.8_2", true, false},
		{"26.7.7", "26.7.8_2", false, false},
		{"27.1", "26.7.8_2", true, false},
		{"26.7.8-beta", "26.7.8_2", false, true},
		{"", "26.7.8_2", false, true},
	}
	for _, row := range rows {
		got, err := trustPackageVersionAtLeast(row.actual, row.minimum)
		if row.fail {
			if err == nil {
				t.Fatalf("%s: expected malformed version", row.actual)
			}
			continue
		}
		if err != nil || got != row.want {
			t.Fatalf("%s: got %v %v, want %v", row.actual, got, err, row.want)
		}
	}
}

func TestTrustCAKeylessRequest(t *testing.T) {
	t.Parallel()
	if trustCAKeylessRequest(map[string]string{"action": "existing", "prv_payload": "key"}, "ca", true, true) {
		t.Fatal("matching key must not require the keyless gate")
	}
	if !trustCAKeylessRequest(map[string]string{"action": "existing", "crt_payload": "cert"}, "ca", true, true) {
		t.Fatal("certificate-only existing import must require the keyless gate")
	}
	if trustCAKeylessRequest(map[string]string{"action": "internal"}, "ca", true, true) {
		t.Fatal("internal create must not require the keyless gate")
	}
	if !trustCAKeylessRequest(map[string]string{"crt_payload": "cert"}, "ca", false, false) {
		t.Fatal("keyless metadata update must require the keyless gate")
	}
	if trustCAKeylessRequest(map[string]string{"action": "existing"}, "cert", true, true) {
		t.Fatal("certificate resources must not use the CA keyless gate")
	}
}
