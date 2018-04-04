package main

// BEFORE RUNNING:
// ---------------
// 1. If not already done, enable the Google Cloud Resource Manager API
//    and check the quota for your project at
//    https://console.developers.google.com/apis/api/cloudresourcemanager
// 2. This sample uses Application Default Credentials for authentication.
//    If not already done, install the gcloud CLI from
//    https://cloud.google.com/sdk/ and run
//    `gcloud beta auth application-default login`.
//    For more information, see
//    https://developers.google.com/identity/protocols/application-default-credentials
// 3. Install and update the Go dependencies by running `go get -u` in the
//    project directory.

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func main() {

	if len(os.Args) < 3 {
		log.Fatalf("usage: gproj org-id billing-account-id")
	}

	verbose := os.Getenv("VERBOSE") != ""

	orgId := os.Args[1]
	billingAccount := strings.ToUpper(os.Args[2])
	log.Printf("organization=[%s] billing=[%s] VERBOSE=%v\n", orgId, billingAccount, verbose)

	ctx := context.Background()

	c, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	cloudresourcemanagerService, err := cloudresourcemanager.New(c)
	if err != nil {
		log.Fatal(err)
	}

	billingService, errBill := cloudbilling.New(c)
	if errBill != nil {
		log.Fatal(errBill)
	}

	var countBillingEmpty int
	var countBillingOk int
	var countBillingOther int
	var countOrgOk int
	var countOrgOther int

	report := func(project, account string) {
		if account == "" {
			countBillingEmpty++
			return // skip empty account
		}
		if account == billingAccount {
			countBillingOk++
			return
		}
		fmt.Printf("wrong billing: %s %s\n", project, account)
		countBillingOther++
	}

	respProj, errProj := cloudresourcemanagerService.Projects.List().Context(ctx).Do()
	if errProj != nil {
		log.Fatal(err)
	}

	log.Printf("full project list: %d", len(respProj.Projects))

	for _, project := range respProj.Projects {
		pid := project.ProjectId

		if verbose {
			log.Printf("verifying project: %s", pid)
		}

		reqBody := &cloudresourcemanager.GetAncestryRequest{}

		resp, err := cloudresourcemanagerService.Projects.GetAncestry(pid, reqBody).Context(ctx).Do()
		if err != nil {
			log.Fatal(err)
		}

		for _, anc := range resp.Ancestor {

			if anc.ResourceId.Type != "organization" {
				continue // not under org
			}

			org := anc.ResourceId.Id

			if verbose {
				log.Printf("verifying project: %s org=[%s]", pid, org)
			}

			if org != orgId {
				countOrgOther++
				continue // wrong org
			}

			countOrgOk++

			proj := "projects/" + pid
			info, err := billingService.Projects.GetBillingInfo(proj).Context(ctx).Do()
			if err != nil {
				log.Fatal(err)
			}

			account := strings.TrimPrefix(info.BillingAccountName, "billingAccounts/")

			if verbose {
				log.Printf("verifying project: %s org=[%s] billing=[%s]", pid, org, account)
			}

			report(pid, account)

			break
		}
	}

	log.Printf("projects for other orgs: %d\n", countOrgOther)
	log.Printf("projects for specified org: %d\n", countOrgOk)
	log.Printf("billing unassigned: %d\n", countBillingEmpty)
	log.Printf("billing ok: %d\n", countBillingOk)
	log.Printf("billing wrong: %d\n", countBillingOther)
}
