package argocd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

const testDataDir = "argocd_testdata"

const payloadAppList = "payload-GET-applications-brief.json"

// const payloadAppRefresh = "payload-GET-application-refresh.json"

// const payloadManagedResources = "payload-GET-managed-resources.json"
// const payloadManifests = "payload-GET-manifest-current.json"
// const payloadManifestsChange = "payload-GET-manifest-change-1.json"
// const payloadError = "payload-GET-manifest-bad-kustomize.json"

func readFileToByteArray(fileName string) ([]byte, string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("Error getting current working directory: %w", err)
	}

	filePath := filepath.Join(workingDir, testDataDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, filePath, fmt.Errorf("error reading file '%s': %w", filePath, err)
	}

	return data, filePath, nil
}

func TestFilterApplications(t *testing.T) {
	var a []v1alpha1.Application

	result, _ := filterApplications(a, "o", "r", "m", "m", "")
	if len(result) != 0 {
		t.Error("Empty param didn't lead to empty result")
	}
	payload, _, err := readFileToByteArray(payloadAppList)
	if err != nil {
		t.Errorf("Failed to read %s: %v", payloadAppList, err)
	}
	var appList v1alpha1.ApplicationList
	if err := json.Unmarshal(payload, &appList); err != nil {
		t.Errorf("Error decoding ApplicationList payload: %v", err)
	}
	a = appList.Items
	if err != nil {
		t.Errorf("decodeApplicationListPayload() failed: %s", err)
	}

	result, _ = filterApplications(a, "o", "r", "m", "m", "")
	if len(result) != 0 {
		t.Error("Unmatchable params didn't lead to empty result")
	}

	result, _ = filterApplications(a, "vince-riv", "argo-diff", "main", "refs/heads/main", "")
	if len(result) != 1 {
		t.Error("Push to main should have matched 1 (auto-sync off)")
	}

	result, _ = filterApplications(a, "vince-riv", "argo-diff", "main", "dev", "main")
	if len(result) != 1 {
		t.Error("Push to dev should have matched")
	}

	result, _ = filterApplications(a, "vince-riv", "argo-diff", "main", "dev", "not_main")
	if len(result) != 0 {
		t.Error("Non-main baseRef should not have matched")
	}

	a[0].Spec.Source.TargetRevision = "main"
	a[1].Spec.Source.TargetRevision = "main"
	result, _ = filterApplications(a, "vince-riv", "argo-diff", "main", "refs/heads/main", "")
	if len(result) != 1 {
		t.Error("Push to main should have matched (targetRev main) (auto-sync off)")
	}

	result, _ = filterApplications(a, "vince-riv", "argo-diff", "main", "dev", "main")
	if len(result) != 1 {
		t.Error("Push to dev should have matched (targetRev main)")
	}

	a[1].Spec.SyncPolicy = &v1alpha1.SyncPolicy{}
	result, _ = filterApplications(a, "vince-riv", "argo-diff", "main", "refs/heads/main", "")
	if len(result) != 1 {
		t.Error("Push to main should have matched (targetRev main) (auto-sync still off)")
	}

	a[1].Spec.SyncPolicy.Automated = &v1alpha1.SyncPolicyAutomated{}
	result, _ = filterApplications(a, "vince-riv", "argo-diff", "main", "refs/heads/main", "")
	if len(result) != 0 {
		t.Error("Push to main should NOT have matched (targetRev main) (auto-sync ENABLED)")
	}
}
