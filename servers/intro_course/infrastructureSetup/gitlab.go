package infrastructureSetup

import (
	"errors"

	"github.com/prompt-edu/prompt/servers/intro_course/infrastructureSetup/data"
	log "github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const IN_PROGRESS_LABEL_ID int64 = 53319
const IN_REVIEW_LABEL_ID int64 = 53320
const ASE_GROUP_ID int64 = 186940

func getClient() (*gitlab.Client, error) {
	// Create a client
	git, err := gitlab.NewClient(InfrastructureServiceSingleton.gitlabAccessToken, gitlab.WithBaseURL("https://gitlab.lrz.de/api/v4"))
	if err != nil {
		log.Error("Failed to create client: ", err)
		return nil, err
	}
	return git, nil

}

func createCourseIterationGroup(courseIteration string, parentID int64) (*gitlab.Group, error) {
	// Create a top level group
	git, err := getClient()
	if err != nil {
		return nil, err
	}

	exists, group, err := checkIfSubGroupExists(courseIteration, parentID)
	if err != nil {
		log.Error("failed to create course iteration group: ", err)
		return nil, err
	}

	if exists {
		return group, nil
	}

	group, _, err = git.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:                  gitlab.Ptr(courseIteration),
		ParentID:              gitlab.Ptr(parentID),
		ProjectCreationLevel:  gitlab.Ptr(gitlab.MaintainerProjectCreation),
		SubGroupCreationLevel: gitlab.Ptr(gitlab.MaintainerSubGroupCreationLevelValue),
		AutoDevopsEnabled:     gitlab.Ptr(false),
		Path:                  gitlab.Ptr(courseIteration),
	})
	if err != nil {
		log.Error("failed to create course iteration group: ", err)
		return nil, err
	}

	return group, nil
}

func createDeveloperTopLevelGroup(parentGroupID int64) (*gitlab.Group, error) {
	return createGitlabGroup(parentGroupID, "developer", gitlab.NoOneProjectCreation, gitlab.OwnerSubGroupCreationLevelValue)
}

// create Groups for tutors and coaches
func createTeachingGroup(parentGroupID int64, groupName string) (*gitlab.Group, error) {
	return createGitlabGroup(parentGroupID, groupName, gitlab.DeveloperProjectCreation, gitlab.OwnerSubGroupCreationLevelValue)
}

func createGitlabGroup(parentGroupID int64, groupName string, projectCreationLevel gitlab.ProjectCreationLevelValue, subGroupCreationLevel gitlab.SubGroupCreationLevelValue) (*gitlab.Group, error) {
	// Create a top level group
	git, err := getClient()
	if err != nil {
		log.Error("failed to create group: ", groupName, " due to failed client creation")
		return nil, err
	}

	exists, group, err := checkIfSubGroupExists(groupName, parentGroupID)
	if err != nil {
		log.Error("failed to create course iteration group: ", err)
		return nil, err
	}

	if exists {
		return group, nil
	}

	// Create a group
	group, _, err = git.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:                  gitlab.Ptr(groupName),
		ParentID:              gitlab.Ptr(parentGroupID),
		ProjectCreationLevel:  gitlab.Ptr(projectCreationLevel),
		SubGroupCreationLevel: gitlab.Ptr(subGroupCreationLevel),
		AutoDevopsEnabled:     gitlab.Ptr(false),
		Path:                  gitlab.Ptr(groupName),
	})

	if err != nil {
		log.Error("failed to create developer group: ", err)
		return nil, err
	}

	return group, nil
}

func getUserID(username string) (*gitlab.User, error) {
	git, err := getClient()
	if err != nil {
		log.Error("failed to get client: ", err)
		return nil, err
	}

	userOpts := &gitlab.ListUsersOptions{
		Username: gitlab.Ptr(username),
	}

	users, _, err := git.Users.ListUsers(userOpts)
	if err != nil {
		log.Error("failed to get user with username : ", username, ", error: ", err)
		return nil, err
	}

	if len(users) != 1 || users[0] == nil {
		log.Error("failed to get user id: user not found")
		return nil, errors.New("user not found")
	}
	return users[0], nil
}

func CreateStudentProject(repoName string, devID, tutorID int64, introCourseID, devGroupID int64, studentName, submissionDeadline string) error {
	git, err := getClient()
	if err != nil {
		log.Error("failed to get client: ", err)
		return errors.New("failed create student project")
	}

	p := &gitlab.CreateProjectOptions{
		Name:                             gitlab.Ptr(repoName),
		NamespaceID:                      gitlab.Ptr(introCourseID),
		SharedRunnersEnabled:             gitlab.Ptr(true),
		OnlyAllowMergeIfPipelineSucceeds: gitlab.Ptr(true),
		BuildsAccessLevel:                gitlab.Ptr(gitlab.PrivateAccessControl),
		ContainerRegistryAccessLevel:     gitlab.Ptr(gitlab.DisabledAccessControl),
		EnvironmentsAccessLevel:          gitlab.Ptr(gitlab.DisabledAccessControl), // disable environments
		FeatureFlagsAccessLevel:          gitlab.Ptr(gitlab.DisabledAccessControl), // disable feature flags
		ForkingAccessLevel:               gitlab.Ptr(gitlab.DisabledAccessControl), // disable forking
		InfrastructureAccessLevel:        gitlab.Ptr(gitlab.DisabledAccessControl), // disable infrastructure
		PackagesEnabled:                  gitlab.Ptr(false),                        // disable packages
		ReleasesAccessLevel:              gitlab.Ptr(gitlab.DisabledAccessControl), // disable releases
		SecurityAndComplianceAccessLevel: gitlab.Ptr(gitlab.DisabledAccessControl), // disable security & compliance
		SnippetsAccessLevel:              gitlab.Ptr(gitlab.DisabledAccessControl), // disable snippets
		WikiAccessLevel:                  gitlab.Ptr(gitlab.DisabledAccessControl), // disable wiki
		RequirementsAccessLevel:          gitlab.Ptr(gitlab.DisabledAccessControl), // disable requirements
		ModelExperimentsAccessLevel:      gitlab.Ptr(gitlab.DisabledAccessControl), // disable model experiments
		ModelRegistryAccessLevel:         gitlab.Ptr(gitlab.DisabledAccessControl), // disable model registry
		PagesAccessLevel:                 gitlab.Ptr(gitlab.DisabledAccessControl), // disable pages
		MonitorAccessLevel:               gitlab.Ptr(gitlab.DisabledAccessControl), // disable monitor
	}

	project, _, err := git.Projects.CreateProject(p)
	if err != nil {
		log.Error("failed to create project: ", err)
		return errors.New("failed create student project")
	}

	err = createProjectFiles(git, project.ID, studentName, submissionDeadline)
	if err != nil {
		return err
	}

	// Add branch protection
	_, _, err = git.ProtectedBranches.ProtectRepositoryBranches(project.ID, &gitlab.ProtectRepositoryBranchesOptions{
		Name:             gitlab.Ptr("main"),
		PushAccessLevel:  gitlab.Ptr(gitlab.MaintainerPermissions),
		MergeAccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	})
	if err != nil {
		log.Error("failed to add branch protect rules: ", err)
		return errors.New("failed add branch protect rules")
	}

	err = createIssueBoard(git, project.ID)
	if err != nil {
		return err
	}

	// Add project members (Last step as this might fail if the tutor is already a member of a "higher" group)
	err = addProjectMembers(git, project.ID, tutorID, devID, devGroupID)
	if err != nil {
		return err
	}

	// Add MR approval rule
	_, _, err = git.Projects.CreateProjectApprovalRule(project.ID, &gitlab.CreateProjectLevelRuleOptions{
		Name:              gitlab.Ptr("Tutor Approval"),
		ApprovalsRequired: gitlab.Ptr(int64(1)),
		UserIDs:           gitlab.Ptr([]int64{tutorID}),
	})
	if err != nil {
		log.Error("failed to add MR approval rule: ", err)
		return errors.New("failed add MR approval rule")
	}

	return nil
}

func addProjectMembers(git *gitlab.Client, projectID, tutorID, devID, devGroupID int64) error {
	// Add student to the project
	_, _, err := git.ProjectMembers.AddProjectMember(projectID, &gitlab.AddProjectMemberOptions{
		UserID:      gitlab.Ptr(devID),
		AccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	})
	if err != nil {
		log.Error("failed to add student to project: ", err)
		return errors.New("failed add student to project")
	}

	// Add student to the developer group
	_, _, err = git.GroupMembers.AddGroupMember(devGroupID, &gitlab.AddGroupMemberOptions{
		UserID:      gitlab.Ptr(devID),
		AccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	})
	if err != nil {
		log.Error("failed to add student to developer group: ", err)
		return errors.New("failed add student to developer group")
	}

	// Add tutor to the project
	_, _, err = git.ProjectMembers.AddProjectMember(projectID, &gitlab.AddProjectMemberOptions{
		UserID:      gitlab.Ptr(tutorID),
		AccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	})
	if err != nil {
		log.Error("failed to add tutor to project: ", err)
		return errors.New("failed add tutor to project")
	}

	return nil
}

func createIssueBoard(git *gitlab.Client, projectID int64) error {
	// Setup issue board
	issueBoard, _, err := git.Boards.CreateIssueBoard(projectID, &gitlab.CreateIssueBoardOptions{
		Name: gitlab.Ptr("Issue Board"),
	})
	if err != nil {
		log.Error("failed to create issue board: ", err)
		return errors.New("failed create issue board")
	}

	// Add issue board lists
	_, _, err = git.Boards.CreateIssueBoardList(projectID, issueBoard.ID, &gitlab.CreateIssueBoardListOptions{
		LabelID: gitlab.Ptr(IN_PROGRESS_LABEL_ID),
	})
	if err != nil {
		log.Error("failed to create issue board list: ", err)
		return errors.New("failed create issue board list")
	}

	_, _, err = git.Boards.CreateIssueBoardList(projectID, issueBoard.ID, &gitlab.CreateIssueBoardListOptions{
		LabelID: gitlab.Ptr(IN_REVIEW_LABEL_ID),
	})
	if err != nil {
		log.Error("failed to create issue board list: ", err)
		return errors.New("failed create issue board list")
	}
	return nil
}

func createProjectFiles(git *gitlab.Client, projectID int64, studentName, submissionDeadline string) error {
	// Add custom README
	_, _, err := git.RepositoryFiles.CreateFile(projectID, "README.md", &gitlab.CreateFileOptions{
		Branch:        gitlab.Ptr("main"),
		Content:       gitlab.Ptr(data.GetReadme(studentName, submissionDeadline)),
		CommitMessage: gitlab.Ptr("Add custom README"),
	})
	if err != nil {
		log.Error("failed to add custom README: ", err)
		return errors.New("failed add custom README")
	}

	// Add custom swiftlint
	_, _, err = git.RepositoryFiles.CreateFile(projectID, ".swiftlint.yml", &gitlab.CreateFileOptions{
		Branch:        gitlab.Ptr("main"),
		Content:       gitlab.Ptr(data.GetSwiftlint()),
		CommitMessage: gitlab.Ptr("Add custom .swiftlint.yml"),
	})
	if err != nil {
		log.Error("failed to add custom .swiftlint.yml: ", err)
		return errors.New("failed add custom .swiftlint.yml")
	}

	// Add custom gitignore
	_, _, err = git.RepositoryFiles.CreateFile(projectID, ".gitignore", &gitlab.CreateFileOptions{
		Branch:        gitlab.Ptr("main"),
		Content:       gitlab.Ptr(data.GetGitignore()),
		CommitMessage: gitlab.Ptr("Add custom .gitignore"),
	})
	if err != nil {
		log.Error("failed to add custom .gitignore: ", err)
		return errors.New("failed add custom .gitignore")
	}

	return nil
}

func getSubGroup(groupName string, parentGroupID int64) (*gitlab.Group, error) {
	git, err := getClient()
	if err != nil {
		log.Error("failed to get group: ", err)
		return nil, err
	}
	groups, _, err := git.Groups.ListSubGroups(parentGroupID, &gitlab.ListSubGroupsOptions{
		AllAvailable: gitlab.Ptr(true),
	})
	if err != nil {
		log.Error("failed to get group: ", err)
		return nil, err
	}

	for _, group := range groups {
		log.Info("group: ", group.Name, " parent: ", group.ParentID)
		if group.Name == groupName && group.ParentID == parentGroupID {
			return group, nil
		}
	}
	return nil, errors.New("subgroup not found")
}

func checkIfSubGroupExists(groupName string, parentGroupID int64) (bool, *gitlab.Group, error) {
	git, err := getClient()
	if err != nil {
		log.Error("failed to get group: ", err)
		return false, nil, err
	}

	groups, _, err := git.Groups.ListSubGroups(parentGroupID, &gitlab.ListSubGroupsOptions{
		Search:       gitlab.Ptr(groupName),
		AllAvailable: gitlab.Ptr(true),
	})
	if err != nil {
		log.Error("failed to get group: ", err)
		return false, nil, err
	}

	for _, group := range groups {
		log.Info("group: ", group.Name, " parent: ", group.ParentID)
		if group.Name == groupName && group.ParentID == parentGroupID {
			return true, group, nil
		}
	}
	return false, nil, nil
}
