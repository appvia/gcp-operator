# Operator to manage GCP projects

## Setup tasks

### Get the details of your GCP organization and billing account (or create them if they dont exist yet...)
```
$ gcloud organizations list
$ gcloud beta billing accounts list
```
### Set some env vars
```
$ export ORG_ID=YOUR_ORG_ID # Get from above
$ export BILLING_ACCOUNT_ID=YOUR_BILLING_ACCOUNT_ID # Get from above
$ export ADMIN_PROJECT_NAME=YOUR_ADMIN_PROJECT_NAME # Set to something sensible
$ export ADMIN_SERVICE_ACCOUNT_NAME=YOUR_ADMIN_SERVICE_ACCOUNT_NAME # Set to something sensible
$ export CREDS_PATH=~/.config/gcloud/konduktor-admin.json
```
### Create an admin project and link to billing
```
$ gcloud projects create ${ADMIN_PROJECT_NAME} \
  --organization ${ORG_ID} \
  --set-as-default
$ gcloud beta billing projects link ${ADMIN_PROJECT_NAME} \
  --billing-account ${BILLING_ACCOUNT_ID}
```
### Create an admin service account
```
$ gcloud iam service-accounts create konduktor \
  --display-name "Konduktor admin account"
$ gcloud iam service-accounts keys create ${CREDS_PATH} \
  --iam-account konduktor@${ADMIN_PROJECT_NAME}.iam.gserviceaccount.com
$ gcloud projects add-iam-policy-binding ${ADMIN_PROJECT_NAME} \
  --member serviceAccount:konduktor@${ADMIN_PROJECT_NAME}.iam.gserviceaccount.com \
  --role roles/viewer
```
### Enable some APIs
```
# gcloud services enable cloudresourcemanager.googleapis.com
# gcloud services enable cloudbilling.googleapis.com
# gcloud services enable iam.googleapis.com
# gcloud services enable compute.googleapis.com
# gcloud services enable serviceusage.googleapis.com
```
### Add org level permissions to allow service account to create projects
```
$ gcloud organizations add-iam-policy-binding ${ORG_ID} \
  --member serviceAccount:konduktor@${ADMIN_PROJECT_NAME}.iam.gserviceaccount.com \
  --role roles/resourcemanager.projectCreator
$ gcloud organizations add-iam-policy-binding ${ORG_ID} \
  --member serviceAccount:konduktor@${ADMIN_PROJECT_NAME}.iam.gserviceaccount.com \
  --role roles/billing.user
```
