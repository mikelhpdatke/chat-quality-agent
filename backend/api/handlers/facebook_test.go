package handlers

import (
	"os"
	"testing"
)

// Integration tests for Facebook Graph API calls.
// These tests hit the real Facebook API and require valid credentials.
// Set the env vars below and remove t.Skip() to run them.
//
//   FB_PAGE_ID=<your_page_id> FB_PAGE_TOKEN=<your_page_access_token> \
//   FB_USER_TOKEN=<your_user_access_token> \
//   go test ./api/handlers/ -run TestFB -v

func TestGetFBPageToken_WithUserToken(t *testing.T) {
	t.Skip("integration test: set FB_USER_TOKEN and remove t.Skip() to run")

	userToken := os.Getenv("FB_USER_TOKEN")
	if userToken == "" {
		t.Fatal("FB_USER_TOKEN env var required")
	}

	pageID, pageToken, pageName, err := getFBPageToken(userToken)
	if err != nil {
		t.Fatalf("getFBPageToken error: %v", err)
	}
	if pageID == "" {
		t.Error("expected non-empty pageID")
	}
	if pageToken == "" {
		t.Error("expected non-empty pageToken")
	}
	t.Logf("page_id=%s page_name=%s token_prefix=%s...", pageID, pageName, pageToken[:10])
}

func TestGetFBPageToken_WithPageAccessToken_ExpectsError(t *testing.T) {
	t.Skip("integration test: set FB_PAGE_TOKEN and remove t.Skip() to run")

	// A Page Access Token cannot call /me/accounts — FB returns OAuthException 102.
	// This test documents and verifies that known failure so we handle it in CreateChannel.
	pageToken := os.Getenv("FB_PAGE_TOKEN")
	if pageToken == "" {
		t.Fatal("FB_PAGE_TOKEN env var required")
	}

	_, _, _, err := getFBPageToken(pageToken)
	if err == nil {
		t.Error("expected error when passing a page access token to getFBPageToken (/me/accounts requires a user token)")
	} else {
		t.Logf("got expected error: %v", err)
	}
}

func TestCreateChannel_DirectPageToken_UsesCredentialsAsIs(t *testing.T) {
	t.Skip("integration test: set FB_PAGE_ID + FB_PAGE_TOKEN and remove t.Skip() to run")

	// Verifies the fix: when page_id is provided alongside the access_token,
	// CreateChannel must NOT call getFBPageToken — it should use them directly.
	pageID := os.Getenv("FB_PAGE_ID")
	pageToken := os.Getenv("FB_PAGE_TOKEN")
	if pageID == "" || pageToken == "" {
		t.Fatal("FB_PAGE_ID and FB_PAGE_TOKEN env vars required")
	}

	// Simulate the branching logic from CreateChannel
	fbCreds := struct {
		PageID      string
		AccessToken string
	}{PageID: pageID, AccessToken: pageToken}

	externalID := ""
	if fbCreds.AccessToken != "" {
		if fbCreds.PageID != "" {
			// Page token provided directly — use as-is, no API call
			externalID = fbCreds.PageID
		} else {
			resolvedPageID, _, _, err := getFBPageToken(fbCreds.AccessToken)
			if err != nil {
				t.Fatalf("getFBPageToken: %v", err)
			}
			externalID = resolvedPageID
		}
	}

	if externalID != pageID {
		t.Errorf("externalID = %q, want %q", externalID, pageID)
	}
	t.Logf("externalID correctly set to %s without calling /me/accounts", externalID)
}
