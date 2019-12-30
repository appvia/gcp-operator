package gcpadminproject

import (
	"context"

	gcpv1alpha1 "github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	core "github.com/appvia/hub-apis/pkg/apis/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

var logger = logf.Log.WithName("controller_gcpadminproject")

// Add creates a new GCPAdminProject Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGCPAdminProject{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gcpadminproject-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GCPAdminProject
	err = c.Watch(&source.Kind{Type: &gcpv1alpha1.GCPAdminProject{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileGCPAdminProject implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGCPAdminProject{}

// ReconcileGCPAdminProject reconciles a GCPAdminProject object
type ReconcileGCPAdminProject struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGCPAdminProject) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := logger.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GCPAdminProject")

	// Fetch the GCPAdminProject instance
	adminProjectInstance := &gcpv1alpha1.GCPAdminProject{}

	if err := r.client.Get(context.TODO(), request.NamespacedName, adminProjectInstance); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	reqLogger.Info("Found the GCPAdminProject")

	adminToken := &gcpv1alpha1.GCPAdminToken{}


	reference := types.NamespacedName{
		Namespace: adminProjectInstance.Spec.Use.Namespace,
		Name:      adminProjectInstance.Spec.Use.Name,
	}

	ctx := context.Background()

	err := r.client.Get(ctx, reference, adminToken)

	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Found the GCPAdminToken")

	bearer := adminToken.Spec.Token

	reqLogger.Info(bearer)

	// Get project details from spec
	projectId, projectName, parentType, parentId := adminProjectInstance.Spec.ProjectId, adminProjectInstance.Spec.ProjectName, adminProjectInstance.Spec.ParentType, adminProjectInstance.Spec.ParentId

	// Check if project already exists
	projectExists, err := HttpProjectExists(ctx, bearer, projectId)

	if err != nil {
		return reconcile.Result{}, err
	}

	if projectExists {
		_, project, err := HttpGetProject(ctx, bearer, projectId)

		billingAccountName, err := HttpGetBilling(projectId, bearer)

		if projectName != project.Name || parentType != project.Parent.Type || parentId != project.Parent.Id {
			// Project exists but state differs
			updateOperationName, err := HttpUpdateProject(ctx, bearer, projectId, projectName, parentId, parentType)

			// Set status to pending
			adminProjectInstance.Status.Status = core.PendingStatus

			if err := r.client.Status().Update(ctx, adminProjectInstance); err != nil {
				logger.Error(err, "failed to update the resource status")
				return reconcile.Result{}, err
			}

			// Wait for project update operation to complete
			_, err = HttpWaitForOperation(updateOperationName, bearer)

			if err != nil {
				return reconcile.Result{}, err
			}
		}

		if billingAccountName != adminProjectInstance.Spec.BillingAccountName {
			// Project exists but billing differs
			err = HttpUpdateBilling(projectId, billingAccountName, bearer)
			if err != nil {
				return reconcile.Result{}, err
			}
			// Set status to pending
			adminProjectInstance.Status.Status = core.PendingStatus

			if err := r.client.Status().Update(ctx, adminProjectInstance); err != nil {
				logger.Error(err, "failed to update the resource status")
				return reconcile.Result{}, err
			}
		}

		// Set status to success
		adminProjectInstance.Status.Status = core.SuccessStatus

		if err := r.client.Status().Update(ctx, adminProjectInstance); err != nil {
			logger.Error(err, "failed to update the resource status")

			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	// Create project
	operationName, err := HttpCreateProject(bearer, projectId, projectName, parentId, parentType)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Set status to pending
	adminProjectInstance.Status.Status = core.PendingStatus

	if err := r.client.Status().Update(ctx, adminProjectInstance); err != nil {
		logger.Error(err, "failed to update the resource status")

		return reconcile.Result{}, err
	}

	// Wait for operation to complete
	_, err = HttpWaitForOperation(operationName, bearer)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Set billing account for admin project
	err = HttpUpdateBilling(projectId, adminProjectInstance.Spec.BillingAccountName, bearer)

	if err != nil {
		return reconcile.Result{}, err
	}

	if err := r.client.Status().Update(ctx, adminProjectInstance); err != nil {
		logger.Error(err, "failed to update the resource status")
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

		operationName, err := HttpEnableAPI(projectId, s, bearer)
		reqLogger.Info("Waiting for operation:", operationName)
		_, err = HttpWaitForOperation(operationName, bearer)

		if err != nil {
			logger.Error(err, "failed to enable the service:", s)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Enabled service:", s)
	}

	// Set status to success
	adminProjectInstance.Status.Status = core.SuccessStatus

	if err := r.client.Status().Update(ctx, adminProjectInstance); err != nil {
		logger.Error(err, "failed to update the resource status")

		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
