package orgdatacore

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Service implements the core organizational data service
type Service struct {
	mu      sync.RWMutex
	data    *Data
	version DataVersion
}

// NewService creates a new organizational data service
func NewService() *Service {
	return &Service{}
}

// LoadFromDataSource loads organizational data from a data source
func (s *Service) LoadFromDataSource(ctx context.Context, source DataSource) error {
	reader, err := source.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load from data source %s: %w", source.String(), err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			fmt.Printf("Warning: failed to close reader: %v\n", closeErr)
		}
	}()

	// Read all data
	// Decode JSON streaming to avoid buffering entire file
	var orgData Data
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&orgData); err != nil {
		return fmt.Errorf("failed to parse JSON from source %s: %w", source.String(), err)
	}

	// Update service data
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = &orgData
	s.version = DataVersion{
		LoadTime:      time.Now(),
		OrgCount:      len(orgData.Lookups.Orgs),
		EmployeeCount: len(orgData.Lookups.Employees),
	}

	return nil
}

// StartDataSourceWatcher starts watching a data source for changes
func (s *Service) StartDataSourceWatcher(ctx context.Context, source DataSource) error {
	// Perform an initial load so data is immediately available
	if err := s.LoadFromDataSource(ctx, source); err != nil {
		return err
	}

	// Start watcher that reloads on change
	callback := func() error {
		return s.LoadFromDataSource(ctx, source)
	}

	return source.Watch(ctx, callback)
}

// GetVersion returns the current data version
func (s *Service) GetVersion() DataVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// GetEmployeeByUID returns an employee by UID
func (s *Service) GetEmployeeByUID(uid string) *Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Employees == nil {
		return nil
	}

	if emp, exists := s.data.Lookups.Employees[uid]; exists {
		return &emp
	}
	return nil
}

// GetEmployeeBySlackID returns an employee by Slack ID
func (s *Service) GetEmployeeBySlackID(slackID string) *Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.SlackIDMappings.SlackUIDToUID == nil || s.data.Lookups.Employees == nil {
		return nil
	}

	uid := s.data.Indexes.SlackIDMappings.SlackUIDToUID[slackID]
	if uid == "" {
		return nil
	}

	if emp, exists := s.data.Lookups.Employees[uid]; exists {
		return &emp
	}
	return nil
}

// GetEmployeeByGitHubID returns an employee by GitHub ID
func (s *Service) GetEmployeeByGitHubID(githubID string) *Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Employees == nil {
		return nil
	}

	uid := s.data.Indexes.GitHubIDMappings.GitHubUIDToUID[githubID]
	if uid == "" {
		return nil
	}

	if emp, exists := s.data.Lookups.Employees[uid]; exists {
		return &emp
	}
	return nil
}

// GetTeamByName returns a team by name
func (s *Service) GetTeamByName(teamName string) *Team {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Teams == nil {
		return nil
	}

	// Get team directly from teams lookup
	team, exists := s.data.Lookups.Teams[teamName]
	if !exists {
		return nil
	}
	return &team
}

// GetOrgByName returns an organization by name
func (s *Service) GetOrgByName(orgName string) *Org {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Orgs == nil {
		return nil
	}

	// Get org directly from orgs lookup
	org, exists := s.data.Lookups.Orgs[orgName]
	if !exists {
		return nil
	}
	return &org
}

// GetTeamsForUID returns all teams a UID is a member of
func (s *Service) GetTeamsForUID(uid string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.Membership.MembershipIndex == nil {
		return []string{}
	}

	memberships := s.data.Indexes.Membership.MembershipIndex[uid]
	var teams []string
	for _, membership := range memberships {
		if membership.Type == MembershipTypeTeam {
			teams = append(teams, membership.Name)
		}
	}
	return teams
}

// GetTeamsForSlackID returns all teams a Slack user is a member of
func (s *Service) GetTeamsForSlackID(slackID string) []string {
	uid := s.getUIDFromSlackID(slackID)
	if uid == "" {
		return []string{}
	}
	return s.GetTeamsForUID(uid)
}

// GetTeamMembers returns all members of a team
func (s *Service) GetTeamMembers(teamName string) []Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Teams == nil {
		return []Employee{}
	}

	// Get team directly from teams lookup
	team, exists := s.data.Lookups.Teams[teamName]
	if !exists {
		return []Employee{}
	}

	// Get employee objects for each UID
	var members []Employee
	for _, uid := range team.Group.ResolvedPeopleUIDList {
		if emp, exists := s.data.Lookups.Employees[uid]; exists {
			members = append(members, emp)
		}
	}

	return members
}

// IsEmployeeInTeam checks if an employee is in a specific team
func (s *Service) IsEmployeeInTeam(uid string, teamName string) bool {
	teams := s.GetTeamsForUID(uid)
	for _, team := range teams {
		if team == teamName {
			return true
		}
	}
	return false
}

// IsSlackUserInTeam checks if a Slack user is in a specific team
func (s *Service) IsSlackUserInTeam(slackID string, teamName string) bool {
	uid := s.getUIDFromSlackID(slackID)
	if uid == "" {
		return false
	}
	return s.IsEmployeeInTeam(uid, teamName)
}

// IsEmployeeInOrg checks if an employee is in a specific organization
func (s *Service) IsEmployeeInOrg(uid string, orgName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.Membership.MembershipIndex == nil {
		return false
	}

	memberships := s.data.Indexes.Membership.MembershipIndex[uid]

	// Get relationship index once
	relationshipIndex := s.data.Indexes.Membership.RelationshipIndex
	teamsIndex := relationshipIndex["teams"]

	for _, membership := range memberships {
		if membership.Type == MembershipTypeOrg && membership.Name == orgName {
			return true
		} else if membership.Type == MembershipTypeTeam {
			// Check if team belongs to the specified org through relationship index
			if teamRelationships, exists := teamsIndex[membership.Name]; exists {
				for _, org := range teamRelationships.Ancestry.Orgs {
					if org == orgName {
						return true
					}
				}
			}
		}
	}
	return false
}

// IsSlackUserInOrg checks if a Slack user is in a specific organization
func (s *Service) IsSlackUserInOrg(slackID string, orgName string) bool {
	uid := s.getUIDFromSlackID(slackID)
	if uid == "" {
		return false
	}
	return s.IsEmployeeInOrg(uid, orgName)
}

// GetUserOrganizations returns the complete organizational hierarchy a Slack user belongs to
func (s *Service) GetUserOrganizations(slackUserID string) []OrgInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.Membership.MembershipIndex == nil {
		return []OrgInfo{}
	}

	uid := s.getUIDFromSlackID(slackUserID)
	if uid == "" {
		return []OrgInfo{}
	}

	memberships := s.data.Indexes.Membership.MembershipIndex[uid]
	var orgs []OrgInfo
	seenItems := make(map[string]bool) // Track seen items to avoid duplicates

	// Get relationship index once
	relationshipIndex := s.data.Indexes.Membership.RelationshipIndex
	teamsIndex := relationshipIndex["teams"]

	for _, membership := range memberships {
		if membership.Type == MembershipTypeOrg {
			// Direct organization membership
			if !seenItems[membership.Name] {
				orgs = append(orgs, OrgInfo{
					Name: membership.Name,
					Type: OrgInfoTypeOrganization,
				})
				seenItems[membership.Name] = true
			}
		} else if membership.Type == MembershipTypeTeam {
			// Add the team membership itself
			if !seenItems[membership.Name] {
				orgs = append(orgs, OrgInfo{
					Name: membership.Name,
					Type: OrgInfoTypeTeam,
				})
				seenItems[membership.Name] = true
			}

			// Get team's hierarchy directly from relationship index
			if teamRelationships, exists := teamsIndex[membership.Name]; exists {
				// Add all ancestry items
				addAncestryItems(&orgs, &seenItems, teamRelationships.Ancestry)
			}
		}
	}

	return orgs
}

// addAncestryItems adds all ancestry items to the orgs slice, avoiding duplicates
func addAncestryItems(orgs *[]OrgInfo, seenItems *map[string]bool, ancestry struct {
	Orgs       []string `json:"orgs"`
	Teams      []string `json:"teams"`
	Pillars    []string `json:"pillars"`
	TeamGroups []string `json:"team_groups"`
}) {
	// Add organizations
	for _, orgName := range ancestry.Orgs {
		if !(*seenItems)[orgName] {
			*orgs = append(*orgs, OrgInfo{
				Name: orgName,
				Type: OrgInfoTypeOrganization,
			})
			(*seenItems)[orgName] = true
		}
	}

	// Add pillars
	for _, pillarName := range ancestry.Pillars {
		if !(*seenItems)[pillarName] {
			*orgs = append(*orgs, OrgInfo{
				Name: pillarName,
				Type: OrgInfoTypePillar,
			})
			(*seenItems)[pillarName] = true
		}
	}

	// Add team groups
	for _, teamGroupName := range ancestry.TeamGroups {
		if !(*seenItems)[teamGroupName] {
			*orgs = append(*orgs, OrgInfo{
				Name: teamGroupName,
				Type: OrgInfoTypeTeamGroup,
			})
			(*seenItems)[teamGroupName] = true
		}
	}

	// Add parent teams
	for _, parentTeamName := range ancestry.Teams {
		if !(*seenItems)[parentTeamName] {
			*orgs = append(*orgs, OrgInfo{
				Name: parentTeamName,
				Type: OrgInfoTypeParentTeam,
			})
			(*seenItems)[parentTeamName] = true
		}
	}
}

// getUIDFromSlackID returns the UID for a given Slack ID
func (s *Service) getUIDFromSlackID(slackID string) string {
	if s.data == nil || s.data.Indexes.SlackIDMappings.SlackUIDToUID == nil {
		return ""
	}
	return s.data.Indexes.SlackIDMappings.SlackUIDToUID[slackID]
}

// GetTeamRepositories returns all repositories for a given team
func (s *Service) GetTeamRepositories(teamName string) []Repository {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Teams == nil {
		return []Repository{}
	}

	team, exists := s.data.Lookups.Teams[teamName]
	if !exists {
		return []Repository{}
	}

	return team.Group.Repos
}

// GetTeamJiraProjects returns all JIRA project configurations for a given team
func (s *Service) GetTeamJiraProjects(teamName string) []JiraConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Teams == nil {
		return []JiraConfig{}
	}

	team, exists := s.data.Lookups.Teams[teamName]
	if !exists {
		return []JiraConfig{}
	}

	return team.Group.Jiras
}

// GetTeamComponents returns all components for a given team
func (s *Service) GetTeamComponents(teamName string) []Component {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Lookups.Teams == nil {
		return []Component{}
	}

	team, exists := s.data.Lookups.Teams[teamName]
	if !exists {
		return []Component{}
	}

	return team.Group.ComponentList
}

// GetTeamByJiraProject returns the team that owns a specific JIRA project
// Returns nil if no team owns the project
func (s *Service) GetTeamByJiraProject(projectKey string) *Team {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.Jira == nil {
		return nil
	}

	projectInfo, exists := s.data.Indexes.Jira[projectKey]
	if !exists || len(projectInfo.ProjectLevel) == 0 {
		return nil
	}

	// Get the first team owner (there's usually only one)
	for _, owner := range projectInfo.ProjectLevel {
		if owner.Type == "team" {
			if team, exists := s.data.Lookups.Teams[owner.Name]; exists {
				return &team
			}
		}
	}

	return nil
}

// GetOrgByJiraProject returns the org that owns a specific JIRA project
// Returns nil if no org owns the project
func (s *Service) GetOrgByJiraProject(projectKey string) *Org {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.Jira == nil {
		return nil
	}

	projectInfo, exists := s.data.Indexes.Jira[projectKey]
	if !exists || len(projectInfo.ProjectLevel) == 0 {
		return nil
	}

	// Get the first org owner
	for _, owner := range projectInfo.ProjectLevel {
		if owner.Type == "org" {
			if org, exists := s.data.Lookups.Orgs[owner.Name]; exists {
				return &org
			}
		}
	}

	return nil
}

// GetJiraProjectOwners returns all owners (teams and orgs) for a JIRA project
func (s *Service) GetJiraProjectOwners(projectKey string) []JiraOwner {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.Jira == nil {
		return []JiraOwner{}
	}

	projectInfo, exists := s.data.Indexes.Jira[projectKey]
	if !exists {
		return []JiraOwner{}
	}

	return projectInfo.ProjectLevel
}

// GetAllJiraProjects returns all JIRA project keys in the index
func (s *Service) GetAllJiraProjects() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil || s.data.Indexes.Jira == nil {
		return []string{}
	}

	projects := make([]string, 0, len(s.data.Indexes.Jira))
	for projectKey := range s.data.Indexes.Jira {
		projects = append(projects, projectKey)
	}

	return projects
}
