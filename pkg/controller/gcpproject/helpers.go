package gcpproject

import (
	"log"
	"time"

	"github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	"golang.org/x/net/context"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	servicemanagement "google.golang.org/api/servicemanagement/v1"
)

// VerifyCredentials is responsible for verifying GCP creds
func VerifyCredentials(credentials *v1alpha1.GCPCredentials) error {
	return nil
}

func GoogleCRMClient(ctx context.Context, key string) (c *cloudresourcemanager.Service, err error) {
	options := []option.ClientOption{option.WithCredentialsJSON([]byte(key))}

	c, err = cloudresourcemanager.NewService(ctx, options...)

	if err != nil {
		return c, err
	}
	return c, nil
}

func GoogleCloudBillingClient(ctx context.Context, key string) (c *cloudbilling.APIService, err error) {
	options := []option.ClientOption{option.WithCredentialsJSON([]byte(key))}

	c, err = cloudbilling.NewService(ctx, options...)

	if err != nil {
		return c, err
	}
	return c, nil
}

func GoogleServiceManagementClient(ctx context.Context, key string) (sm *servicemanagement.APIService, err error) {
	options := []option.ClientOption{option.WithCredentialsJSON([]byte(key))}

	sm, err = servicemanagement.NewService(ctx, options...)

	if err != nil {
		return sm, err
	}
	return sm, nil
}

func GoogleIAMClient(ctx context.Context, key string) (i *iam.Service, err error) {
	options := []option.ClientOption{option.WithCredentialsJSON([]byte(key))}

	i, err = iam.NewService(ctx, options...)

	if err != nil {
		return i, err
	}
	return i, nil
}

func ProjectExists(ctx context.Context, crm *cloudresourcemanager.Service, projectId string) (exists bool, err error) {

	if err != nil {
		return exists, err
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
		return err
	}

	_, err = crm.Projects.Delete(projectId).Context(ctx).Do()

	return err
}

func CreateProject(ctx context.Context, crm *cloudresourcemanager.Service, projectId, projectName, parentId, parentType string) (operationName string, err error) {

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
		return operationName, err
	}

	return resp.Name, nil
}

func UpdateProject(ctx context.Context, crm *cloudresourcemanager.Service, projectId, projectName, parentId, parentType string) (operationName string, err error) {

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
		return operationName, err
	}
	return resp.Name, nil
}

func WaitForOperationCRM(ctx context.Context, crm *cloudresourcemanager.Service, operationName string) (complete bool, err error) {

	log.Println("Waiting for operation", operationName)

	for {
		resp, err := crm.Operations.Get(operationName).Context(ctx).Do()
		if err != nil {
			return complete, err
		}

		if resp.Done == true {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
	log.Println("Operation complete")
	return
}

func WaitForOperationSM(ctx context.Context, sm *servicemanagement.APIService, operationName string) (complete bool, err error) {

	log.Println("Waiting for operation", operationName)

	for {
		resp, err := sm.Operations.Get(operationName).Context(ctx).Do()
		if err != nil {
			return false, err
		}
		if resp.Done == true {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
	log.Println("Operation complete")
	return
}

func GetProjectBilling(ctx context.Context, cb *cloudbilling.APIService, projectId string) (billingInfo *cloudbilling.ProjectBillingInfo, err error) {
	billingInfo, err = cb.Projects.GetBillingInfo(projectId).Context(ctx).Do()

	if err != nil {
		return billingInfo, err
	}
	return billingInfo, nil
}

func UpdateProjectBilling(ctx context.Context, cb *cloudbilling.APIService, billingAccountName string, projectId string) (err error) {
	name := "projects/" + projectId

	_, err = cb.Projects.UpdateBillingInfo(name, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: "billingAccounts/" + billingAccountName,
		BillingEnabled:     true,
	}).Context(ctx).Do()

	if err != nil {
		return err
	}
	return
}

func EnableAPI(ctx context.Context, sm *servicemanagement.APIService, projectId, serviceName string) (operationName string, err error) {
	log.Println("Enabling service", serviceName, "for project", projectId)

	resp, err := sm.Services.Enable(serviceName, &servicemanagement.EnableServiceRequest{
		ConsumerId: "project:" + projectId,
	}).Context(ctx).Do()

	if err != nil {
		return operationName, err
	}
	return resp.Name, err
}

func CreateServiceAccount(ctx context.Context, i *iam.Service, projectId, name, displayName string) (*iam.ServiceAccount, error) {
	request := &iam.CreateServiceAccountRequest{
		AccountId: name,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: displayName,
		},
	}
	account, err := i.Projects.ServiceAccounts.Create("projects/"+projectId, request).Do()
	if err != nil {
		return nil, err
	}
	log.Printf("Created service account: %v", account)
	return account, nil
}

func CreateServiceAccountKey(ctx context.Context, i *iam.Service, projectId, serviceAccountName string) (key string, err error) {
	resource := "projects/" + projectId + "/serviceAccounts/" + serviceAccountName
	request := &iam.CreateServiceAccountKeyRequest{}
	serviceAccount, err := i.Projects.ServiceAccounts.Keys.Create(resource, request).Do()
	if err != nil {
		return key, err
	}
	log.Printf("Created key: %v", serviceAccountName)
	return serviceAccount.PrivateKeyData, nil
}
