package gcpproject

import (
	"context"

	gcpv1alpha1 "github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	core "github.com/appvia/hub-apis/pkg/apis/core/v1"
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

	err := r.client.Get(context.TODO(), request.NamespacedName, projectInstance)
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

	ctx := context.Background()

	// TODO: get key string from GCPCredentials and use JWTConfigFromJSON
	// https://github.com/golang/oauth2/blob/master/google/google.go#L80

	c, err := GoogleClient(ctx)

	// Check if project already exists
	projectExists, err := ProjectExists(ctx, c, projectId)

	if err != nil {
		return reconcile.Result{}, err
	}

	if projectExists {
		project, err := GetProject(ctx, c, projectId)

		if projectName == project.ProjectName && parentType == project.Parent.Type && parentId == project.Parent.Id {
			// Exists and state as desired
			return reconcile.Result{}, nil
		}

		// Exists but state differs
		updateOperationName, err := UpdateProject(ctx, c, projectId, projectName, parentId, parentType)

		// Set status to pending
		p.project.Status.Status = core.PendingStatus

		// Wait for operation to complete
		_, err = WaitForOperation(ctx, c, operationName)

		if err != nil {
			return reconcile.Result{}, err
		}

		// Set status to success
		p.project.Status.Status = core.SuccessStatus

		return reconcile.Result{}, nil
	}

	// Create project
	operationName, err := CreateProject(ctx, c, projectId, projectName, parentId, parentType)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Set status to pending
	p.project.Status.Status = core.PendingStatus

	// Wait for operation to complete
	_, err = WaitForOperation(ctx, c, operationName)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Set status to success
	p.project.Status.Status = core.SuccessStatus

	return reconcile.Result{}, nil
}
