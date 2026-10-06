package utils

import "testing"

// TestKnownPortsFixedEntries locks the fixed Kubernetes-related port map so a
// rename or refactor cannot silently drop or relabel an entry.
func TestKnownPortsFixedEntries(t *testing.T) {
	want := map[int]string{
		443:   "Kubernetes API port",
		2379:  "Etcd",
		6666:  "Etcd",
		4194:  "cAdvisor for containers metrics",
		6443:  "Kubernetes API port",
		8443:  "Minikube API port",
		8080:  "Insecure API port",
		10250: "Kubelet API anonymous port",
		10255: "Kubelet API read only",
		10256: "Kube proxy health check server",
		9099:  "Calico health check server",
		6782:  "Weave metrics/endpoints",
		6783:  "Weave metrics/endpoints",
		6784:  "Weave metrics/endpoints",
		44134: "Tiller service listening",
	}
	for port, label := range want {
		if got := KnownPorts[port]; got != label {
			t.Errorf("KnownPorts[%d] = %q, want %q", port, got, label)
		}
	}
}

// TestKnownPortsNodePortRange verifies that init() populated the full NodePort
// range and nothing immediately outside it.
func TestKnownPortsNodePortRange(t *testing.T) {
	const label = "Possible NodePort service"
	for _, port := range []int{30000, 31234, 32767} {
		if got := KnownPorts[port]; got != label {
			t.Errorf("KnownPorts[%d] = %q, want %q", port, got, label)
		}
	}
	for _, port := range []int{29999, 32768} {
		if got := KnownPorts[port]; got == label {
			t.Errorf("KnownPorts[%d] unexpectedly labelled %q", port, label)
		}
	}
}
