package gcpproject

import (
	"log"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func GoogleClient(ctx context.Context) (*http.Client, error) {
	// TODO: take cred string bytes as arg
	client, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)

	if err != nil {
		log.Fatal(err)
	}

	return client, nil
}

func ProjectExists(ctx context.Context, c *http.Client, projectId string) (bool, error) {
	cloudresourcemanagerService, err := cloudresourcemanager.New(c)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listing projects matching filter id:" + projectId)

	resp, err := cloudresourcemanagerService.Projects.List().Filter("id:" + projectId).Do()

	if err != nil {
		return false, err
	}

	projectsReturned := len(resp.Projects)

	if projectsReturned == 0 {
		log.Println("Project not found")
		return false, nil
	}

	log.Println("Project found")

	return true, nil
}

func GetProject(ctx context.Context, c *http.Client, projectId string) (project *cloudresourcemanager.Project, err error) {
	cloudresourcemanagerService, err := cloudresourcemanager.New(c)

	project, err = cloudresourcemanagerService.Projects.Get(projectId).Context(ctx).Do()

	return project, nil
}

func DeleteProject(ctx context.Context, c *http.Client, projectId string) error {
	cloudresourcemanagerService, err := cloudresourcemanager.New(c)

	if err != nil {
		log.Fatal(err)
	}

	_, err = cloudresourcemanagerService.Projects.Delete(projectId).Context(ctx).Do()

	return err
}

func CreateProject(ctx context.Context, c *http.Client, projectId, projectName, parentId, parentType string) (operationName string, err error) {
	cloudresourcemanagerService, err := cloudresourcemanager.New(c)

	if err != nil {
		log.Fatal(err)
	}

	parent := &cloudresourcemanager.ResourceId{
		Id:   parentId,
		Type: parentType,
	}

	rb := &cloudresourcemanager.Project{
		Name:      projectName,
		ProjectId: projectId,
		Parent:    parent,
	}

	log.Println("Creating project")

	resp, err := cloudresourcemanagerService.Projects.Create(rb).Context(ctx).Do()

	if err != nil {
		log.Fatal(err)
	}

	return resp.Name, nil
}

func UpdateProject(ctx context.Context, c *http.Client, projectId, projectName, parentId, parentType string) (operationName string, err error) {
	cloudresourcemanagerService, err := cloudresourcemanager.New(c)

	if err != nil {
		log.Fatal(err)
	}

	parent := &cloudresourcemanager.ResourceId{
		Id:   parentId,
		Type: parentType,
	}

	rb := &cloudresourcemanager.Project{
		Name:      projectName,
		ProjectId: projectId,
		Parent:    parent,
	}

	resp, err := cloudresourcemanagerService.Projects.Update(projectId, rb).Context(ctx).Do()

	if err != nil {
		log.Fatal(err)
	}
	return resp.Name, nil
}

func WaitForOperation(ctx context.Context, c *http.Client, operationName string) (complete bool, err error) {
	cloudresourcemanagerService, err := cloudresourcemanager.New(c)

	log.Println("Waiting for operation", operationName)

	if err != nil {
		log.Fatal(err)
	}

	for {
		resp, err := cloudresourcemanagerService.Operations.Get(operationName).Context(ctx).Do()
		if err != nil {
			log.Fatal(err)
		}

		if resp.Done == true {
			break
		}
	}
	log.Println("Operation complete")
	return
}
