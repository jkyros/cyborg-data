package orgdatacore

import (
	"testing"
)

func TestGetEntityByStableID(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name         string
		stableID     string
		expectFound  bool
		expectedName string
		expectedType string
	}{
		{
			name:         "team stable ID",
			stableID:     "a1b2c3d4",
			expectFound:  true,
			expectedName: "test-team",
			expectedType: "team",
		},
		{
			name:         "org stable ID",
			stableID:     "c9d0e1f2",
			expectFound:  true,
			expectedName: "test-org",
			expectedType: "org",
		},
		{
			name:         "pillar stable ID",
			stableID:     "e7f8a9b0",
			expectFound:  true,
			expectedName: "engineering",
			expectedType: "pillar",
		},
		{
			name:         "team group stable ID",
			stableID:     "c1d2e3f4",
			expectFound:  true,
			expectedName: "backend-teams",
			expectedType: "team_group",
		},
		{
			name:        "nonexistent stable ID",
			stableID:    "00000000",
			expectFound: false,
		},
		{
			name:        "empty stable ID",
			stableID:    "",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetEntityByStableID(tt.stableID)
			if tt.expectFound {
				if result == nil {
					t.Fatalf("GetEntityByStableID(%q) = nil, expected result", tt.stableID)
				}
				if result.Name != tt.expectedName {
					t.Errorf("GetEntityByStableID(%q).Name = %q, expected %q", tt.stableID, result.Name, tt.expectedName)
				}
				if result.Type != tt.expectedType {
					t.Errorf("GetEntityByStableID(%q).Type = %q, expected %q", tt.stableID, result.Type, tt.expectedType)
				}
			} else {
				if result != nil {
					t.Errorf("GetEntityByStableID(%q) = %+v, expected nil", tt.stableID, result)
				}
			}
		})
	}
}

func TestGetTeamByStableID(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name         string
		stableID     string
		expectFound  bool
		expectedName string
	}{
		{
			name:         "valid team stable ID",
			stableID:     "a1b2c3d4",
			expectFound:  true,
			expectedName: "test-team",
		},
		{
			name:        "org stable ID returns nil for team lookup",
			stableID:    "c9d0e1f2",
			expectFound: false,
		},
		{
			name:        "nonexistent stable ID",
			stableID:    "00000000",
			expectFound: false,
		},
		{
			name:        "empty stable ID",
			stableID:    "",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamByStableID(tt.stableID)
			if tt.expectFound {
				if result == nil {
					t.Fatalf("GetTeamByStableID(%q) = nil, expected result", tt.stableID)
				}
				if result.Name != tt.expectedName {
					t.Errorf("GetTeamByStableID(%q).Name = %q, expected %q", tt.stableID, result.Name, tt.expectedName)
				}
			} else {
				if result != nil {
					t.Errorf("GetTeamByStableID(%q) = %+v, expected nil", tt.stableID, result)
				}
			}
		})
	}
}

func TestGetOrgByStableID(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name         string
		stableID     string
		expectFound  bool
		expectedName string
	}{
		{
			name:         "valid org stable ID",
			stableID:     "c9d0e1f2",
			expectFound:  true,
			expectedName: "test-org",
		},
		{
			name:        "team stable ID returns nil for org lookup",
			stableID:    "a1b2c3d4",
			expectFound: false,
		},
		{
			name:        "nonexistent stable ID",
			stableID:    "00000000",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetOrgByStableID(tt.stableID)
			if tt.expectFound {
				if result == nil {
					t.Fatalf("GetOrgByStableID(%q) = nil, expected result", tt.stableID)
				}
				if result.Name != tt.expectedName {
					t.Errorf("GetOrgByStableID(%q).Name = %q, expected %q", tt.stableID, result.Name, tt.expectedName)
				}
			} else {
				if result != nil {
					t.Errorf("GetOrgByStableID(%q) = %+v, expected nil", tt.stableID, result)
				}
			}
		})
	}
}

func TestGetPillarByStableID(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name         string
		stableID     string
		expectFound  bool
		expectedName string
	}{
		{
			name:         "valid pillar stable ID",
			stableID:     "e7f8a9b0",
			expectFound:  true,
			expectedName: "engineering",
		},
		{
			name:        "team stable ID returns nil for pillar lookup",
			stableID:    "a1b2c3d4",
			expectFound: false,
		},
		{
			name:        "nonexistent stable ID",
			stableID:    "00000000",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetPillarByStableID(tt.stableID)
			if tt.expectFound {
				if result == nil {
					t.Fatalf("GetPillarByStableID(%q) = nil, expected result", tt.stableID)
				}
				if result.Name != tt.expectedName {
					t.Errorf("GetPillarByStableID(%q).Name = %q, expected %q", tt.stableID, result.Name, tt.expectedName)
				}
			} else {
				if result != nil {
					t.Errorf("GetPillarByStableID(%q) = %+v, expected nil", tt.stableID, result)
				}
			}
		})
	}
}

func TestGetTeamGroupByStableID(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name         string
		stableID     string
		expectFound  bool
		expectedName string
	}{
		{
			name:         "valid team group stable ID",
			stableID:     "c1d2e3f4",
			expectFound:  true,
			expectedName: "backend-teams",
		},
		{
			name:        "team stable ID returns nil for team group lookup",
			stableID:    "a1b2c3d4",
			expectFound: false,
		},
		{
			name:        "nonexistent stable ID",
			stableID:    "00000000",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamGroupByStableID(tt.stableID)
			if tt.expectFound {
				if result == nil {
					t.Fatalf("GetTeamGroupByStableID(%q) = nil, expected result", tt.stableID)
				}
				if result.Name != tt.expectedName {
					t.Errorf("GetTeamGroupByStableID(%q).Name = %q, expected %q", tt.stableID, result.Name, tt.expectedName)
				}
			} else {
				if result != nil {
					t.Errorf("GetTeamGroupByStableID(%q) = %+v, expected nil", tt.stableID, result)
				}
			}
		})
	}
}
