package gcpadminproject

import (
	"context"

	gcpv1alpha1 "github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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

var log = logf.Log.WithName("controller_gcpadminproject")

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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner GCPAdminProject
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gcpv1alpha1.GCPAdminProject{},
	})
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
	reqLogger.Info("Reconciling GCPProject")

	// Fetch the GCPProject instance
	projectInstance := &gcpv1alpha1.GCPProject{}

	credentials := &gcpv1alpha1.GCPCredentials{}

	reference := types.NamespacedName{
		Namespace: projectInstance.Spec.Use.Namespace,
		Name:      projectInstance.Spec.Use.Name,
	}

	ctx := context.Background()

	err := r.client.Get(ctx, reference, credentials)

	if err != nil {
		log.Fatal(err)
	}

	// Authenticate to cloudresourcemanager
	crm, err := GoogleClient(ctx, credentials.Spec.Key)

	if err != nil {
		log.Fatal(err)
	}

	// Attempt to retrieve the project
	err = r.client.Get(context.TODO(), request.NamespacedName, projectInstance)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Get project details from spec
	projectId, projectName, parentType, parentId := projectInstance.Spec.ProjectId, projectInstance.Spec.ProjectName, projectInstance.Spec.ParentType, projectInstance.Spec.ParentId

	// Check if project already exists
	projectExists, err := ProjectExists(ctx, crm, projectId)

	if err != nil {
		return reconcile.Result{}, err
	}

	if projectExists {
		project, err := GetProject(ctx, crm, projectId)

		if projectName == project.Name && parentType == project.Parent.Type && parentId == project.Parent.Id {
			// Exists and state as desired
			return reconcile.Result{}, nil
		}

		// Exists but state differs
		updateOperationName, err := UpdateProject(ctx, crm, projectId, projectName, parentId, parentType)

		// Set status to pending
		projectInstance.Status.Status = core.PendingStatus

		if err := r.client.Status().Update(ctx, projectInstance); err != nil {
			logger.Error(err, "failed to update the resource status")

			return reconcile.Result{}, err
		}

		// Wait for operation to complete
		_, err = HttWaitForOperation(ctx, crm, updateOperationName)

		if err != nil {
			return reconcile.Result{}, err
		}

		// Set status to success
		projectInstance.Status.Status = core.SuccessStatus

		if err := r.client.Status().Update(ctx, projectInstance); err != nil {
			logger.Error(err, "failed to update the resource status")

			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	// Create project
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
	_, err = WaitForOperation(ctx, crm, operationName)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Set status to success
	projectInstance.Status.Status = core.SuccessStatus

	if err := r.client.Status().Update(ctx, projectInstance); err != nil {
		logger.Error(err, "failed to update the resource status")

		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}