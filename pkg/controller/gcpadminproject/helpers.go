package gcpadminproject

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	iam "google.golang.org/api/iam/v1"
	servicemanagement "google.golang.org/api/servicemanagement/v1"
)

// VerifyCredentials is responsible for verifying GCP creds
func VerifyCredentials(credentials *v1alpha1.GCPCredentials) error {
	return nil
}

func CallGoogleRest(bearer, url, method string, reqBody []byte) (responseBody []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Add("Authorization", "Bearer "+bearer)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if os.Getenv("DEBUG") == "true" {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		fmt.Println("response Body:", string(body))
	}
	return body, err
}

func HttpCreateProject(bearer, projectId, projectName, parentId, parentType string) (operationName string, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/projects/"
	parent := &cloudresourcemanager.ResourceId{
		Id:   parentId,
		Type: parentType,
	}
	project := &cloudresourcemanager.Project{
		Name:      projectName,
		ProjectId: projectId,
		Parent:    parent,
	}
	reqBody, err := json.Marshal(project)
	resBody, err := CallGoogleRest(bearer, url, "POST", reqBody)

	var operation cloudresourcemanager.Operation

	json.Unmarshal(resBody, &operation)

	operationName = operation.Name

	return operationName, err
}

func HttpUpdateProject(ctx context.Context, bearer, projectId, projectName, parentId, parentType string) (operationName string, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/projects/"
	parent := &cloudresourcemanager.ResourceId{
		Id:   parentId,
		Type: parentType,
	}
	project := &cloudresourcemanager.Project{
		Name:      projectName,
		ProjectId: projectId,
		Parent:    parent,
	}
	reqBody, err := json.Marshal(project)
	resBody, err := CallGoogleRest(bearer, url, "PUT", reqBody)

	var operation cloudresourcemanager.Operation

	json.Unmarshal(resBody, &operation)

	operationName = operation.Name

	return operationName, err
}

func HttpProjectExists(ctx context.Context, bearer, projectId string) (exists bool, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/projects?filter=id:" + projectId

	type projectListResponse struct {
		Projects []cloudresourcemanager.Project `json:"projects"`
	}

	var projects projectListResponse

	resp, err := CallGoogleRest(bearer, url, "GET", make([]byte, 0))

	if err != nil {
		return exists, err
	}

	json.Unmarshal(resp, &projects)

	projectsReturned := len(projects.Projects)

	if projectsReturned == 0 {
		return false, nil
	}

	return true, nil
}

func HttpGetProject(ctx context.Context, bearer, projectId string) (exists bool, project *cloudresourcemanager.Project, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/projects/" + projectId

	if err != nil {
		return
	}

	resp, err := CallGoogleRest(bearer, url, "GET", make([]byte, 0))

	if err != nil {
		return
	}

	json.Unmarshal(resp, &project)

	if err != nil {
		exists = true
	} else {
		exists = false
	}

	return exists, project, err
}

func HttpCreateServiceAccount(bearer, projectId, serviceAccountName, displayName string) (serviceAccount iam.ServiceAccount, err error) {
	url := "https://iam.googleapis.com/v1/projects/" + projectId + "/serviceAccounts"
	serviceAccountRequest := &iam.CreateServiceAccountRequest{
		AccountId: serviceAccountName,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: displayName,
		},
	}
	reqBody, err := json.Marshal(serviceAccountRequest)
	fmt.Println("Creating service account", serviceAccountName, "in project", projectId)
	resBody, err := CallGoogleRest(bearer, url, "POST", reqBody)
	err = json.Unmarshal(resBody, &serviceAccount)
	return serviceAccount, err
}

func HttpGetServiceAccount(bearer, projectId, serviceAccountName string) (serviceAccount iam.ServiceAccount, err error) {
	url := "https://iam.googleapis.com/v1/projects/" + projectId + "/serviceAccounts/" + serviceAccountName
	fmt.Println("Retrieving service account", serviceAccountName, "in project", projectId)
	respBody, err := CallGoogleRest(bearer, url, "GET", make([]byte, 0))
	err = json.Unmarshal(respBody, &serviceAccount)
	return serviceAccount, err
}

func HttpCreateServiceAccountKey(bearer, projectId, serviceAccountName string) (key string, err error) {
	url := "https://iam.googleapis.com/v1/projects/" + projectId + "/serviceAccounts/" + serviceAccountName + "@" + projectId + ".iam.gserviceaccount.com/keys"
	resBody, err := CallGoogleRest(bearer, url, "POST", make([]byte, 0)) // TODO: do this better
	var serviceAccountKey iam.ServiceAccountKey
	err = json.Unmarshal(resBody, &serviceAccountKey)
	return serviceAccountKey.PrivateKeyData, err
}

func dedupePolicy(policy cloudresourcemanager.Policy) (uniquePolicy cloudresourcemanager.Policy) {
	for _, b := range policy.Bindings {
		// Append to new policy if not in already
		if !bindingInPolicy(b, uniquePolicy) {
			uniquePolicy.Bindings = append(uniquePolicy.Bindings, b)
		}
	}
	return uniquePolicy
}

func bindingInPolicy(binding *cloudresourcemanager.Binding, policy cloudresourcemanager.Policy) bool {
	for _, b := range policy.Bindings {
		if reflect.DeepEqual(binding.Members, b.Members) && binding.Condition == b.Condition && binding.Role == b.Role {
			// Binding in policy
			return true
		}
	}
	// Binding not in policy
	return false
}

func HttpGetProjectIam(bearer, projectId string) (policy cloudresourcemanager.Policy, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/projects/" + projectId + ":getIamPolicy"
	resBody, err := CallGoogleRest(bearer, url, "POST", make([]byte, 0))
	json.Unmarshal(resBody, &policy)
	return policy, err
}

func HttpSetProjectIam(bearer, serviceAccountEmail, projectId string) (err error) {
	// Get existing policy and append new required bindings
	existingPolicy, err := HttpGetProjectIam(bearer, projectId)
	if err != nil {
		return err
	}
	binding := &cloudresourcemanager.Binding{
		Members: []string{"serviceAccount:" + serviceAccountEmail},
		Role:    "roles/viewer",
	}
	allBindings := append(existingPolicy.Bindings, binding)
	updatedPolicy := cloudresourcemanager.Policy{
		Bindings: allBindings,
	}
	finalPolicy := dedupePolicy(updatedPolicy)
	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: &finalPolicy,
	}
	reqBody, err := json.Marshal(setIamPolicyRequest)
	if err != nil {
		return err
	}
	url := "https://cloudresourcemanager.googleapis.com/v1/projects/" + projectId + ":setIamPolicy"
	_, err = CallGoogleRest(bearer, url, "POST", reqBody)
	return err
}

func HttpGetOrgIam(bearer, orgId string) (policy cloudresourcemanager.Policy, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/organizations/" + orgId + ":getIamPolicy"
	resBody, err := CallGoogleRest(bearer, url, "POST", make([]byte, 0))
	json.Unmarshal(resBody, &policy)
	return policy, err
}

func HttpSetOrgIam(bearer, serviceAccountEmail, orgId string) (err error) {
	// Get existing policy and append new required bindings
	existingPolicy, err := HttpGetProjectIam(bearer, orgId)
	if err != nil {
		return err
	}
	billingBinding := &cloudresourcemanager.Binding{
		Members: []string{"serviceAccount:" + serviceAccountEmail},
		Role:    "roles/billing.user",
	}
	projectCreatorBinding := &cloudresourcemanager.Binding{
		Members: []string{"serviceAccount:" + serviceAccountEmail},
		Role:    "roles/resourcemanager.projectCreator",
	}
	allBindings := append(existingPolicy.Bindings, billingBinding, projectCreatorBinding)
	updatedPolicy := cloudresourcemanager.Policy{
		Bindings: allBindings,
	}
	finalPolicy := dedupePolicy(updatedPolicy)
	setIamPolicyRequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: &finalPolicy,
	}
	reqBody, err := json.Marshal(setIamPolicyRequest)
	if err != nil {
		return err
	}
	url := "https://cloudresourcemanager.googleapis.com/v1/organizations/" + orgId + ":setIamPolicy"
	_, err = CallGoogleRest(bearer, url, "POST", reqBody)
	return err
}

func HttpWaitForSMOperation(operationName, bearer string) (complete bool, err error) {
	url := "https://servicemanagement.googleapis.com/v1/" + operationName
	for {
		resBody, err := CallGoogleRest(bearer, url, "GET", make([]byte, 0)) // TODO: do this better
		if err != nil {
			return false, err
		}
		var operation servicemanagement.Operation
		json.Unmarshal(resBody, &operation)
		if err != nil {
			return false, err
		}
		if operation.Done {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

	return true, err
}

func HttpWaitForCRMOperation(operationName, bearer string) (complete bool, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/" + operationName
	for {
		resBody, err := CallGoogleRest(bearer, url, "GET", make([]byte, 0)) // TODO: do this better
		if err != nil {
			return false, err
		}
		var operation cloudresourcemanager.Operation
		json.Unmarshal(resBody, &operation)
		if err != nil {
			return false, err
		}
		if operation.Done {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

	return true, err
}

func HttpGetBilling(projectId, bearer string) (billingAccountName string, err error) {
	url := "https://cloudbilling.googleapis.com/v1/projects/" + projectId + "/billingInfo"

	var billingInfo cloudbilling.ProjectBillingInfo

	resp, err := CallGoogleRest(bearer, url, "GET", make([]byte, 0))

	err = json.Unmarshal(resp, &billingInfo)

	if err != nil {
		return billingAccountName, err
	}

	return billingInfo.BillingAccountName, err
}

func HttpUpdateBilling(projectId, billingAccountName, bearer string) (err error) {
	url := "https://cloudbilling.googleapis.com/v1/projects/" + projectId + "/billingInfo"

	billingInfo := &cloudbilling.ProjectBillingInfo{
		BillingAccountName: "billingAccounts/" + billingAccountName,
		BillingEnabled:     true,
	}
	reqBody, err := json.Marshal(billingInfo)
	_, err = CallGoogleRest(bearer, url, "PUT", reqBody)
	return err
}

func HttpEnableAPI(projectId, serviceName, bearer string) (operationName string, err error) {
	url := "https://servicemanagement.googleapis.com/v1/services/" + serviceName + ":enable"

	type consumerBody struct {
		ConsumerId string `json:"consumerId"`
	}

	consumer := &consumerBody{
		ConsumerId: "project:" + projectId,
	}

	reqBody, err := json.Marshal(consumer)

	if err != nil {
		return operationName, err
	}

	resp, err := CallGoogleRest(bearer, url, "POST", reqBody)

	if err != nil {
		return operationName, err
	}

	operation := &cloudresourcemanager.Operation{}

	return operation.Name, json.Unmarshal(resp, operation)
}
