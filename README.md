# gproj
List GCP project billing

Install
=======

    go get golang.org/x/oauth2/google
    go get google.golang.org/api/cloudbilling/v1
    go get google.golang.org/api/cloudresourcemanager/v1
    go get github.com/udhos/gproj
    go install github.com/udhos/gproj

Authentication
==============

You can use this command to provide the SDK with an auth token for your GCP user:

    gcloud auth application-default login

Usage
=====

    gproj org-id billing-account-id
	    
