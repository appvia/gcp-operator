# Operator to manage GCP projects

## Overview
This operator can be used to setup GCP projects linked to an existing organisation and billing account. It will provision a service account within each project, generate credentials and saved them to a custom resource within the cluster.

This operator contains the following CRDs:
* `GCPCredentials` - this defines a GCP service account credential which can be used to authenticate in the GCP SDK/CLI
* `GCPAdminProject` - this defines a GCP admin project, used to create an admin service account which has organisation level permissions, required for creating new projects
* `GCPProject` - this defines a GCP project, intended for use by a team

## Using the operator

### Get the details of your GCP organization and billing account (or [create them](https://cloud.google.com/resource-manager/docs/creating-managing-organization) if they dont exist yet...)
```
$ gcloud organizations list # Note the `ID`
$ gcloud beta billing accounts list # Note the `ACCOUNT_ID`
```
### Get a bearer token from your account
```
$ gcloud auth application-default print-access-token
```
### Create an `GCPAdminProject` custom resource, using the details above
```
$ kubectl create -f ...
```
### Create a `GCPProject` custom resource, passing a reference to the `GCPCredentials` object
```
$ kubectl create -f ...
```
### Get the project service account credentials to setup gcloud CLI
```
$ kubectl get ... -o yaml
```
### [Activate service account](https://cloud.google.com/sdk/gcloud/reference/auth/activate-service-account) and use the project
```
$ gcloud auth activate-service-account [ACCOUNT] --key-file=./service-account.json
```
