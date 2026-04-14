package orgdatacore

import (
	"reflect"
	"testing"
)

// TestGetTeamRepositories tests fetching repositories for a team
func TestGetTeamRepositories(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name     string
		teamName string
		expected []Repository
	}{
		{
			name:     "test-team with multiple repos",
			teamName: "test-team",
			expected: []Repository{
				{
					RepoName:    "https://github.com/example/test-repo",
					Description: "Main test repository",
				},
				{
					RepoName:    "https://github.com/example/test-tools",
					Description: "Testing tools and utilities",
				},
			},
		},
		{
			name:     "platform-team with one repo",
			teamName: "platform-team",
			expected: []Repository{
				{
					RepoName:    "https://github.com/example/platform-core",
					Description: "Core platform services",
				},
			},
		},
		{
			name:     "nonexistent team",
			teamName: "nonexistent-team",
			expected: []Repository{},
		},
		{
			name:     "empty team name",
			teamName: "",
			expected: []Repository{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamRepositories(tt.teamName)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetTeamRepositories(%q) = %v, want %v", tt.teamName, result, tt.expected)
			}
		})
	}
}

// TestGetTeamByRepository tests reverse lookup from repository to team
func TestGetTeamByRepository(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name         string
		repoURL      string
		expectedTeam *string // use pointer so we can test nil
	}{
		{
			name:         "test-repo belongs to test-team",
			repoURL:      "https://github.com/example/test-repo",
			expectedTeam: strPtr("test-team"),
		},
		{
			name:         "test-tools belongs to test-team",
			repoURL:      "https://github.com/example/test-tools",
			expectedTeam: strPtr("test-team"),
		},
		{
			name:         "platform-core belongs to platform-team",
			repoURL:      "https://github.com/example/platform-core",
			expectedTeam: strPtr("platform-team"),
		},
		{
			name:         "nonexistent repository",
			repoURL:      "https://github.com/example/nonexistent",
			expectedTeam: nil,
		},
		{
			name:         "empty repository URL",
			repoURL:      "",
			expectedTeam: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamByRepository(tt.repoURL)
			if tt.expectedTeam == nil {
				if result != nil {
					t.Errorf("GetTeamByRepository(%q) = %v, want nil", tt.repoURL, result.Name)
				}
			} else {
				if result == nil {
					t.Errorf("GetTeamByRepository(%q) = nil, want team %s", tt.repoURL, *tt.expectedTeam)
				} else if result.Name != *tt.expectedTeam {
					t.Errorf("GetTeamByRepository(%q) = %v, want %v", tt.repoURL, result.Name, *tt.expectedTeam)
				}
			}
		})
	}
}

// TestGetTeamsByRepositoryPattern tests pattern-based repository search
func TestGetTeamsByRepositoryPattern(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		pattern       string
		expectedTeams []string
	}{
		{
			name:          "pattern 'test' matches test-team repos",
			pattern:       "test",
			expectedTeams: []string{"test-team"},
		},
		{
			name:          "pattern 'platform' matches platform-team",
			pattern:       "platform",
			expectedTeams: []string{"platform-team"},
		},
		{
			name:          "pattern 'example' matches all teams",
			pattern:       "example",
			expectedTeams: []string{"test-team", "platform-team"},
		},
		{
			name:          "pattern 'nonexistent' matches no teams",
			pattern:       "nonexistent-pattern",
			expectedTeams: []string{},
		},
		{
			name:          "empty pattern",
			pattern:       "",
			expectedTeams: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamsByRepositoryPattern(tt.pattern)

			// Extract team names from results
			var teamNames []string
			for _, team := range result {
				teamNames = append(teamNames, team.Name)
			}

			if len(tt.expectedTeams) == 0 && len(teamNames) == 0 {
				return // Both empty, test passes
			}

			// Check that all expected teams are present
			for _, expected := range tt.expectedTeams {
				found := false
				for _, name := range teamNames {
					if name == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetTeamsByRepositoryPattern(%q) missing expected team %q, got %v", tt.pattern, expected, teamNames)
				}
			}
		})
	}
}

// TestGetTeamJiraProjects tests fetching JIRA projects for a team
func TestGetTeamJiraProjects(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		teamName      string
		expectedCount int
		checkProject  string // Check if this project exists in results
	}{
		{
			name:          "test-team has JIRA projects",
			teamName:      "test-team",
			expectedCount: 4, // 2 project configs + 2 dashboards
			checkProject:  "TEST",
		},
		{
			name:          "platform-team has JIRA projects",
			teamName:      "platform-team",
			expectedCount: 2, // 1 project + 1 dashboard
			checkProject:  "PLAT",
		},
		{
			name:          "nonexistent team",
			teamName:      "nonexistent-team",
			expectedCount: 0,
			checkProject:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamJiraProjects(tt.teamName)
			if len(result) != tt.expectedCount {
				t.Errorf("GetTeamJiraProjects(%q) returned %d projects, want %d", tt.teamName, len(result), tt.expectedCount)
			}

			if tt.checkProject != "" {
				found := false
				for _, jira := range result {
					if jira.Project == tt.checkProject {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetTeamJiraProjects(%q) did not include project %q", tt.teamName, tt.checkProject)
				}
			}
		})
	}
}

// TestGetTeamsByJiraProject tests reverse lookup from JIRA project to teams
func TestGetTeamsByJiraProject(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		projectKey    string
		expectedTeams []string
	}{
		{
			name:          "TEST project belongs to test-team",
			projectKey:    "TEST",
			expectedTeams: []string{"test-team"},
		},
		{
			name:          "PLAT project belongs to platform-team",
			projectKey:    "PLAT",
			expectedTeams: []string{"platform-team"},
		},
		{
			name:          "NOPROJECT (no _project_level) falls back to component owner",
			projectKey:    "NOPROJECT",
			expectedTeams: []string{"test-team"},
		},
		{
			name:          "nonexistent project",
			projectKey:    "NONEXIST",
			expectedTeams: []string{},
		},
		{
			name:          "empty project key",
			projectKey:    "",
			expectedTeams: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamsByJiraProject(tt.projectKey)

			if len(result) != len(tt.expectedTeams) {
				t.Errorf("GetTeamsByJiraProject(%q) returned %d teams, want %d", tt.projectKey, len(result), len(tt.expectedTeams))
				return
			}

			// Check that all expected teams are present
			for _, expectedTeam := range tt.expectedTeams {
				found := false
				for _, team := range result {
					if team.Name == expectedTeam {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetTeamsByJiraProject(%q) missing expected team %q", tt.projectKey, expectedTeam)
				}
			}
		})
	}
}

// TestGetAllJiraProjects tests listing all JIRA projects
func TestGetAllJiraProjects(t *testing.T) {
	service := setupTestService(t)

	result := service.GetAllJiraProjects()

	// Should have TEST and PLAT projects
	if len(result) != 2 {
		t.Errorf("GetAllJiraProjects() returned %d projects, want 2", len(result))
	}

	// Check that both expected projects are present
	expectedProjects := map[string]bool{"TEST": false, "PLAT": false}
	for _, project := range result {
		if _, exists := expectedProjects[project]; exists {
			expectedProjects[project] = true
		}
	}

	for project, found := range expectedProjects {
		if !found {
			t.Errorf("GetAllJiraProjects() did not include expected project %q", project)
		}
	}
}

// TestGetTeamJiraDashboards tests fetching JIRA dashboards for a team
func TestGetTeamJiraDashboards(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		teamName      string
		expectedCount int
		checkType     string // Check if this dashboard type exists
	}{
		{
			name:          "test-team has dashboards",
			teamName:      "test-team",
			expectedCount: 2, // scrum-dashboard and roadmap-dashboard
			checkType:     "scrum-dashboard",
		},
		{
			name:          "platform-team has dashboard",
			teamName:      "platform-team",
			expectedCount: 1, // workload-dashboard
			checkType:     "workload-dashboard",
		},
		{
			name:          "nonexistent team",
			teamName:      "nonexistent-team",
			expectedCount: 0,
			checkType:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamJiraDashboards(tt.teamName)
			if len(result) != tt.expectedCount {
				t.Errorf("GetTeamJiraDashboards(%q) returned %d dashboards, want %d", tt.teamName, len(result), tt.expectedCount)
			}

			if tt.checkType != "" {
				found := false
				for _, dashboard := range result {
					for _, dashboardType := range dashboard.Types {
						if dashboardType == tt.checkType {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					t.Errorf("GetTeamJiraDashboards(%q) did not include dashboard type %q", tt.teamName, tt.checkType)
				}
			}
		})
	}
}

// TestGetTeamJiraDashboardByType tests fetching a specific dashboard type
func TestGetTeamJiraDashboardByType(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		teamName      string
		dashboardType string
		shouldFind    bool
		expectedView  string
	}{
		{
			name:          "test-team has scrum-dashboard",
			teamName:      "test-team",
			dashboardType: "scrum-dashboard",
			shouldFind:    true,
			expectedView:  "https://issues.example.com/secure/RapidBoard.jspa?rapidView=123",
		},
		{
			name:          "test-team has roadmap-dashboard",
			teamName:      "test-team",
			dashboardType: "roadmap-dashboard",
			shouldFind:    true,
			expectedView:  "https://issues.example.com/secure/PortfolioPlanView.jspa?id=456",
		},
		{
			name:          "platform-team has workload-dashboard",
			teamName:      "platform-team",
			dashboardType: "workload-dashboard",
			shouldFind:    true,
			expectedView:  "https://issues.example.com/secure/Dashboard.jspa?selectPageId=789",
		},
		{
			name:          "test-team does not have workload-dashboard",
			teamName:      "test-team",
			dashboardType: "workload-dashboard",
			shouldFind:    false,
			expectedView:  "",
		},
		{
			name:          "nonexistent team",
			teamName:      "nonexistent-team",
			dashboardType: "scrum-dashboard",
			shouldFind:    false,
			expectedView:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamJiraDashboardByType(tt.teamName, tt.dashboardType)
			if tt.shouldFind {
				if result == nil {
					t.Errorf("GetTeamJiraDashboardByType(%q, %q) = nil, want dashboard", tt.teamName, tt.dashboardType)
				} else if result.View != tt.expectedView {
					t.Errorf("GetTeamJiraDashboardByType(%q, %q) view = %q, want %q", tt.teamName, tt.dashboardType, result.View, tt.expectedView)
				}
			} else {
				if result != nil {
					t.Errorf("GetTeamJiraDashboardByType(%q, %q) = %v, want nil", tt.teamName, tt.dashboardType, result)
				}
			}
		})
	}
}

// TestGetTeamComponents tests fetching team components
func TestGetTeamComponents(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		teamName      string
		expectedCount int
	}{
		{
			name:          "test-team has components",
			teamName:      "test-team",
			expectedCount: 1,
		},
		{
			name:          "platform-team has no components",
			teamName:      "platform-team",
			expectedCount: 0,
		},
		{
			name:          "nonexistent team",
			teamName:      "nonexistent-team",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamComponents(tt.teamName)
			if len(result) != tt.expectedCount {
				t.Errorf("GetTeamComponents(%q) returned %d components, want %d", tt.teamName, len(result), tt.expectedCount)
			}
		})
	}
}

// TestGetTeamsByComponent tests reverse lookup from component to teams
func TestGetTeamsByComponent(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		componentName string
		expectedTeams []string
	}{
		{
			name:          "backend component belongs to test-team",
			componentName: "backend",
			expectedTeams: []string{"test-team"},
		},
		{
			name:          "nonexistent component",
			componentName: "nonexistent-component",
			expectedTeams: []string{},
		},
		{
			name:          "empty component name",
			componentName: "",
			expectedTeams: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTeamsByComponent(tt.componentName)

			if len(result) != len(tt.expectedTeams) {
				t.Errorf("GetTeamsByComponent(%q) returned %d teams, want %d", tt.componentName, len(result), len(tt.expectedTeams))
				return
			}

			// Check that all expected teams are present
			for _, expectedTeam := range tt.expectedTeams {
				found := false
				for _, team := range result {
					if team.Name == expectedTeam {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetTeamsByComponent(%q) missing expected team %q", tt.componentName, expectedTeam)
				}
			}
		})
	}
}

// TestGetJiraProjectOwners tests fetching owners of a JIRA project
func TestGetJiraProjectOwners(t *testing.T) {
	service := setupTestService(t)

	tests := []struct {
		name          string
		projectKey    string
		expectedCount int
		expectedName  string
		expectedType  string
	}{
		{
			name:          "TEST project has one owner",
			projectKey:    "TEST",
			expectedCount: 1,
			expectedName:  "test-team",
			expectedType:  "team",
		},
		{
			name:          "PLAT project has one owner",
			projectKey:    "PLAT",
			expectedCount: 1,
			expectedName:  "platform-team",
			expectedType:  "team",
		},
		{
			name:          "nonexistent project",
			projectKey:    "NONEXIST",
			expectedCount: 0,
			expectedName:  "",
			expectedType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetJiraProjectOwners(tt.projectKey)
			if len(result) != tt.expectedCount {
				t.Errorf("GetJiraProjectOwners(%q) returned %d owners, want %d", tt.projectKey, len(result), tt.expectedCount)
			}

			if tt.expectedCount > 0 && len(result) > 0 {
				if result[0].Name != tt.expectedName {
					t.Errorf("GetJiraProjectOwners(%q) owner name = %q, want %q", tt.projectKey, result[0].Name, tt.expectedName)
				}
				if result[0].Type != tt.expectedType {
					t.Errorf("GetJiraProjectOwners(%q) owner type = %q, want %q", tt.projectKey, result[0].Type, tt.expectedType)
				}
			}
		})
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
