package nodemeta

import (
	"testing"
)

func TestGetNodeReturnsNilForMissing(t *testing.T) {
	s := &service{nodes: make(map[string]*NodeSnapshot)}
	if got := s.getNode("nonexistent"); got != nil {
		t.Errorf("expected nil for missing node, got %+v", got)
	}
}

func TestGetNodeReturnsExisting(t *testing.T) {
	s := &service{nodes: make(map[string]*NodeSnapshot)}
	s.nodes["node-1"] = &NodeSnapshot{NodeID: "node-1", MaxMvmNum: 500}
	got := s.getNode("node-1")
	if got == nil {
		t.Fatal("expected non-nil for existing node")
	}
	if got.MaxMvmNum != 500 {
		t.Errorf("expected MaxMvmNum=500, got %d", got.MaxMvmNum)
	}
}

func TestResolveMaxMvmNum(t *testing.T) {
	tests := []struct {
		name          string
		existing      *NodeSnapshot
		reqMaxMvmNum  int64
		wantMaxMvmNum int64
	}{
		{
			name:          "preserve existing when request is zero",
			existing:      &NodeSnapshot{MaxMvmNum: 500},
			reqMaxMvmNum:  0,
			wantMaxMvmNum: 500,
		},
		{
			name:          "preserve existing when request is negative",
			existing:      &NodeSnapshot{MaxMvmNum: 500},
			reqMaxMvmNum:  -1,
			wantMaxMvmNum: 500,
		},
		{
			name:          "use request when request has value",
			existing:      &NodeSnapshot{MaxMvmNum: 500},
			reqMaxMvmNum:  300,
			wantMaxMvmNum: 300,
		},
		{
			name:          "use request when no existing node",
			existing:      nil,
			reqMaxMvmNum:  200,
			wantMaxMvmNum: 200,
		},
		{
			name:          "use request when existing has no limit",
			existing:      &NodeSnapshot{MaxMvmNum: 0},
			reqMaxMvmNum:  200,
			wantMaxMvmNum: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveMaxMvmNum(tt.existing, tt.reqMaxMvmNum)
			if got != tt.wantMaxMvmNum {
				t.Errorf("resolveMaxMvmNum() = %d, want %d", got, tt.wantMaxMvmNum)
			}
		})
	}
}
