package gcpadminproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	iam "google.golang.org/api/iam/v1"
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
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return body, err
}

func HttpCreateProject(bearer, projectId, projectName, orgId string) (operationName string, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/projects/"
	parent := &cloudresourcemanager.ResourceId{
		Id:   orgId,
		Type: "organization",
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

func HttpCreateServiceAccount(bearer, projectId, serviceAccountName, displayName string) (err error) {
	url := "https://iam.googleapis.com/v1/projects/" + projectId + "/serviceAccounts"
	serviceAccount := &iam.CreateServiceAccountRequest{
		AccountId: serviceAccountName,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: displayName,
		},
	}
	reqBody, err := json.Marshal(serviceAccount)
	fmt.Println("Creating service account", serviceAccountName, "in project", projectId)
	_, err = CallGoogleRest(bearer, url, "POST", reqBody)
	return err
}

func HttpCreateServiceAccountKey(bearer, projectId, serviceAccountName string) (key string, err error) {
	url := "https://iam.googleapis.com/v1/projects/" + projectId + "/serviceAccounts/" + serviceAccountName + "@" + projectId + ".iam.gserviceaccount.com/keys"
	fmt.Println("Creating service account key for", serviceAccountName, "in project", projectId)
	resBody, err := CallGoogleRest(bearer, url, "POST", make([]byte, 0)) // TODO: do this better
	var serviceAccountKey iam.ServiceAccountKey
	err = json.Unmarshal(resBody, &serviceAccountKey)
	return serviceAccountKey.PrivateKeyData, err
}

func HttpSetProjectIam(bearer, serviceAccountEmail, projectId string) (err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/projects/" + projectId + ":setIamPolicy"
	binding := &cloudresourcemanager.Binding{
		Members: []string{serviceAccountEmail},
		Role:    "roles/viewer",
	}
	policy := &cloudresourcemanager.Policy{
		Bindings: []*cloudresourcemanager.Binding{binding},
	}
	reqBody, err := json.Marshal(policy)
	_, err = CallGoogleRest(bearer, url, "POST", reqBody)
	return err
}

func HttpSetOrgIam(bearer, serviceAccountEmail, orgId string) (err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/organizations/" + orgId + ":setIamPolicy"
	billingBinding := &cloudresourcemanager.Binding{
		Members: []string{serviceAccountEmail},
		Role:    "roles/billing.user",
	}
	projectCreatorBinding := &cloudresourcemanager.Binding{
		Members: []string{serviceAccountEmail},
		Role:    "roles/resourcemanager.projectCreator",
	}
	policy := &cloudresourcemanager.Policy{
		Bindings: []*cloudresourcemanager.Binding{billingBinding, projectCreatorBinding},
	}
	reqBody, err := json.Marshal(policy)
	log.Println("Setting org policy roles/billing.user and roles/resourcemanager.projectCreator")
	_, err = CallGoogleRest(bearer, url, "POST", reqBody)
	return err
}

func HttpWaitForOperation(operationName, bearer string) (complete bool, err error) {
	url := "https://cloudresourcemanager.googleapis.com/v1/" + operationName
	for {
		log.Println("Checking the status of operation", operationName)
		resBody, err := CallGoogleRest(bearer, url, "GET", make([]byte, 0)) // TODO: do this better
		if err != nil {
			log.Println("Exiting due to remote err")
			log.Fatal(err)
			break
		}
		var operation cloudresourcemanager.Operation
		json.Unmarshal(resBody, &operation)
		if operation.Done {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return true, err
}

func HttpUpdateBilling(projectId, billingAccountName, bearer string) (complete bool, err error) {
	url := "https://cloudbilling.googleapis.com/v1/projects/" + projectId + "/billingInfo"

	billingInfo := &cloudbilling.ProjectBillingInfo{
		BillingAccountName: "billingAccounts/" + billingAccountName,
		BillingEnabled:     true,
	}
	reqBody, err := json.Marshal(billingInfo)
	log.Println("Updating billing account for", projectName, "to", billingAccountName)
	_, err = CallGoogleRest(bearer, url, "PUT", reqBody)
	return err
}