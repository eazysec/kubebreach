package utils

import "testing"

// TestConfigDefaults locks the default flag values so the rename does not alter
// the tool's out-of-the-box behavior.
func TestConfigDefaults(t *testing.T) {
	if AnonRce {
		t.Errorf("AnonRce default = %v, want false", AnonRce)
	}
	if ScanNode {
		t.Errorf("ScanNode default = %v, want false", ScanNode)
	}
	if ScanEtcd {
		t.Errorf("ScanEtcd default = %v, want false", ScanEtcd)
	}
	const wantCmd = "cat /var/run/secrets/kubernetes.io/serviceaccount/token"
	if RceCommand != wantCmd {
		t.Errorf("RceCommand default = %q, want %q", RceCommand, wantCmd)
	}
}
