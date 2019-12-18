package gcpproject

import (
	"log"

	"github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	"golang.org/x/net/context"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/servicemanagement/v1"
)

// VerifyCredentials is responsible for verifying GCP creds
func VerifyCredentials(credentials *v1alpha1.GCPCredentials) error {
	return nil
}

func GoogleCRMClient(ctx context.Context, key string) (c *cloudresourcemanager.Service, err error) {
	options := []option.ClientOption{option.WithCredentialsJSON([]byte(key))}

	c, err = cloudresourcemanager.NewService(ctx, options...)

	if err != nil {
		log.Fatal(err)
	}
	return c, nil
}

func GoogleCloudBillingClient(ctx context.Context, key string) (c *cloudbilling.APIService, err error) {
	options := []option.ClientOption{option.WithCredentialsJSON([]byte(key))}

	c, err = cloudbilling.NewService(ctx, options...)

	if err != nil {
		log.Fatal(err)
	}
	return c, nil
}

func GoogleServiceManagementClient(ctx context.Context, key string) (c *servicemanagement.APIService, err error) {
	options := []option.ClientOption{option.WithCredentialsJSON([]byte(key))}

	c, err = servicemanagement.NewService(ctx, options...)

	if err != nil {
		log.Fatal(err)
	}
	return c, nil
}

func ProjectExists(ctx context.Context, crm *cloudresourcemanager.Service, projectId string) (exists bool, err error) {

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listing projects matching filter id:" + projectId)

	resp, err := crm.Projects.List().Filter("id:" + projectId).Do()

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

func GetProject(ctx context.Context, crm *cloudresourcemanager.Service, projectId string) (project *cloudresourcemanager.Project, err error) {

	project, err = crm.Projects.Get(projectId).Context(ctx).Do()

	return project, nil
}

func DeleteProject(ctx context.Context, crm *cloudresourcemanager.Service, projectId string) (err error) {

	if err != nil {
		log.Fatal(err)
	}

	_, err = crm.Projects.Delete(projectId).Context(ctx).Do()

	return err
}

func CreateProject(ctx context.Context, crm *cloudresourcemanager.Service, projectId, projectName, parentId, parentType string) (operationName string, err error) {

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

	resp, err := crm.Projects.Create(rb).Context(ctx).Do()

	if err != nil {
		log.Fatal(err)
	}

	return resp.Name, nil
}

func UpdateProject(ctx context.Context, crm *cloudresourcemanager.Service, projectId, projectName, parentId, parentType string) (operationName string, err error) {

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

	resp, err := crm.Projects.Update(projectId, rb).Context(ctx).Do()

	if err != nil {
		log.Fatal(err)
	}
	return resp.Name, nil
}

func WaitForOperation(ctx context.Context, crm *cloudresourcemanager.Service, operationName string) (complete bool, err error) {

	log.Println("Waiting for operation", operationName)

	if err != nil {
		log.Fatal(err)
	}

	for {
		resp, err := crm.Operations.Get(operationName).Context(ctx).Do()
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

func UpdateProjectBilling(ctx context.Context, cb *cloudbilling.APIService, billingAccountName string, projectId string) (err error) {
	name := "projects/" + projectId

	_, err = cb.Projects.UpdateBillingInfo(name, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: "billingAccounts/" + billingAccountName,
		BillingEnabled:     true,
	}).Context(ctx).Do()

	if err != nil {
		log.Fatal(err)
	}
	return
}

func EnableAPI(ctx context.Context, sm *servicemanagement.APIService, projectId, serviceName string) (operationName string, err error) {
	log.Println("Enabling service", serviceName, "for project", projectId)

	resp, err := sm.Services.Enable(serviceName, &servicemanagement.EnableServiceRequest{
		ConsumerId: "project:" + projectId,
	}).Context(ctx).Do()

	if err != nil {
		log.Fatal(err)
	}
	operationName = resp.Name
	return operationName, err
}
