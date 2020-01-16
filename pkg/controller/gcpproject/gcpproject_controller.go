package gcpproject

import (
	"context"
	"encoding/base64"

	gcpv1alpha1 "github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	core "github.com/appvia/hub-apis/pkg/apis/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logger = logf.Log.WithName("controller_gcpproject")

// Add creates a new GCPProject Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGCPProject{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gcpproject-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GCPProject
	err = c.Watch(&source.Kind{Type: &gcpv1alpha1.GCPProject{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileGCPProject implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGCPProject{}

// ReconcileGCPProject reconciles a GCPProject object
type ReconcileGCPProject struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GCPProject object and makes changes based on the state read
// and what is in the GCPProject.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGCPProject) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := logger.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GCPProject")

	// Fetch the GCPProject instance
	projectInstance := &gcpv1alpha1.GCPProject{}

	if err := r.client.Get(context.TODO(), request.NamespacedName, projectInstance); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	reqLogger.Info("Found GCPProject CR")

	credentials := &gcpv1alpha1.GCPCredentials{}

	reference := types.NamespacedName{
		Namespace: projectInstance.Spec.Use.Namespace,
		Name:      projectInstance.Spec.Use.Name,
	}

	ctx := context.Background()

	err := r.client.Get(ctx, reference, credentials)

	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Found GCPCredentials CR")

	decoded, err := base64.StdEncoding.DecodeString(credentials.Spec.Key)

	keyString := string(decoded)

	// Authenticate to cloudresourcemanager
	crm, err := GoogleCRMClient(ctx, keyString)

	if err != nil {
		logger.Error(err, "Failed to obtain CRM client")
		return reconcile.Result{}, err
	}

	reqLogger.Info("Authenticated to CRM")

	// Authenticate to cloudbilling
	cb, err := GoogleCloudBillingClient(ctx, keyString)

	if err != nil {
		logger.Error(err, "Failed to obtain CB client")
		return reconcile.Result{}, err
	}

	reqLogger.Info("Authenticated to CB")

	// Get project details from spec
	projectId, projectName, parentType, parentId := projectInstance.Spec.ProjectId, projectInstance.Spec.ProjectName, projectInstance.Spec.ParentType, projectInstance.Spec.ParentId

	// Check if project already exists
	projectExists, err := ProjectExists(ctx, crm, projectId)

	if err != nil {
		return reconcile.Result{}, err
	}

	if projectExists {
		project, err := GetProject(ctx, crm, projectId)

		billingAccount, err := GetProjectBilling(ctx, cb, projectId)

		if projectName != project.Name || parentType != project.Parent.Type || parentId != project.Parent.Id {
			// Exists but state differs
			updateOperationName, err := UpdateProject(ctx, crm, projectId, projectName, parentId, parentType)

			// Set status to pending
			projectInstance.Status.Status = core.PendingStatus

			if err := r.client.Status().Update(ctx, projectInstance); err != nil {
				logger.Error(err, "failed to update the resource status")

				return reconcile.Result{}, err
			}

			// Wait for operation to complete
			_, err = WaitForOperationCRM(ctx, crm, updateOperationName)

			if err != nil {
				return reconcile.Result{}, err
			}

			if err := r.client.Status().Update(ctx, projectInstance); err != nil {
				logger.Error(err, "failed to update the resource status")

				return reconcile.Result{}, err
			}
		} else {
			reqLogger.Info("Project exists and state matches")
		}

		if billingAccount.Name != "billingAccounts/"+projectInstance.Spec.BillingAccountName {
			reqLogger.Info("Project exists but billing account doesnt match, updating")

			err = UpdateProjectBilling(ctx, cb, projectInstance.Spec.BillingAccountName, projectId)

			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Project billing updated")
		} else {
			reqLogger.Info("Project exists and billing account matches")
		}

		// Set status to success
		projectInstance.Status.Status = core.SuccessStatus

		if err := r.client.Status().Update(ctx, projectInstance); err != nil {
			logger.Error(err, "failed to update the resource status")
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	// Project doesnt exist yet, create it
	operationName, err := CreateProject(ctx, crm, projectId, projectName, parentId, parentType)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Set status to pending
	projectInstance.Status.Status = core.PendingStatus

	if err := r.client.Status().Update(ctx, projectInstance); err != nil {
		logger.Error(err, "failed to update the resource status")

		return reconcile.Result{}, err
	}

	// Wait for operation to complete
	_, err = WaitForOperationCRM(ctx, crm, operationName)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Get a service management client for enabling required APIs
	sm, err := GoogleServiceManagementClient(ctx, credentials.Spec.Key)

	if err != nil {
		return reconcile.Result{}, err
	}

	servicesToEnable := []string{
		"cloudresourcemanager.googleapis.com",
		"cloudbilling.googleapis.com",
		"iam.googleapis.com",
		"compute.googleapis.com",
		"serviceusage.googleapis.com",
	}

	// Enable each API in the new project
	for _, s := range servicesToEnable {
		err := func() error {
			name, err := EnableAPI(ctx, sm, projectId, s)
			if err != nil {
				return err
			}
			reqLogger.Info("Waiting for operation:", name)
			if _, err = WaitForOperationSM(ctx, sm, name); err != nil {
				return err
			}
			reqLogger.Info("Enabled service:", s)

			return nil
		}()
		if err != nil {
			logger.Error(err, "failed to enable the service:", err)

			return reconcile.Result{}, err
		}
	}

	// Set billing
	err = UpdateProjectBilling(ctx, cb, projectInstance.Spec.BillingAccountName, projectId)

	if err := r.client.Status().Update(ctx, projectInstance); err != nil {
		logger.Error(err, "failed to update the resource status")

		return reconcile.Result{}, err
	}

	iam, err := GoogleIAMClient(ctx, keyString)

	// Create service account and then create a key
	reqLogger.Info("Creating service account: " + projectInstance.Spec.ServiceAccountName)
	serviceAccount, err := CreateServiceAccount(ctx, iam, projectId, projectInstance.Spec.ServiceAccountName, "Created by the Appvia Hub")

	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Creating service account key for: " + projectInstance.Spec.ServiceAccountName)
	key, err := CreateServiceAccountKey(ctx, iam, projectId, projectInstance.Spec.ServiceAccountName)

	if err != nil {
		return reconcile.Result{}, err
	}

	err = MakeProjectAdmin(ctx, crm, projectId, serviceAccount.Name, serviceAccount.Email)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Create the credential as a CR
	reqLogger.Info("Creating the GCPCredentials CR:" + projectId + "-gcpcreds in namespace: " + request.Namespace)

	serviceAccountCredential := &gcpv1alpha1.GCPCredentials{
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectId + "-gcpcreds",
			Namespace: request.Namespace,
		},
		Spec: gcpv1alpha1.GCPCredentialsSpec{
			Key:            key,
			ProjectId:      projectId,
			OrganizationId: parentId,
		},
		Status: gcpv1alpha1.GCPCredentialsStatus{
			Status: "Success",
		},
	}

	err = r.client.Create(ctx, serviceAccountCredential)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Set status to success
	projectInstance.Status.Status = core.SuccessStatus

	return reconcile.Result{}, nil
}
