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

	orgId := os.Args[1]
	billingAccount := strings.ToUpper(os.Args[2])
	log.Printf("organization=[%s] billing=[%s]\n", orgId, billingAccount)

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

	report := func(project, account string) {
		if account == "" {
			countBillingEmpty++
			return // skip empty account
		}
		account = strings.TrimPrefix(account, "billingAccounts/")
		if account == billingAccount {
			countBillingOk++
			return
		}
		fmt.Printf("%s %s\n", project, account)
		countBillingOther++
	}

	req := cloudresourcemanagerService.Projects.List()
	if err := req.Pages(ctx, func(page *cloudresourcemanager.ListProjectsResponse) error {

		log.Printf("projects: %d", len(page.Projects))

		for _, project := range page.Projects {
			pid := project.ProjectId

			log.Printf("verifying project: %s", pid)

			reqBody := &cloudresourcemanager.GetAncestryRequest{}

			resp, err := cloudresourcemanagerService.Projects.GetAncestry(pid, reqBody).Context(ctx).Do()
			if err != nil {
				log.Fatal(err)
			}

			for _, anc := range resp.Ancestor {
				if anc.ResourceId.Type != "organization" || anc.ResourceId.Id != orgId {
					continue // skip wrong org
				}
				proj := "projects/" + pid
				info, err := billingService.Projects.GetBillingInfo(proj).Context(ctx).Do()
				if err != nil {
					log.Fatal(err)
				}

				report(pid, info.BillingAccountName)

				break
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	log.Printf("billing unassigned: %d\n", countBillingEmpty)
	log.Printf("billing ok: %d\n", countBillingOk)
	log.Printf("billing wrong: %d\n", countBillingOther)
}
