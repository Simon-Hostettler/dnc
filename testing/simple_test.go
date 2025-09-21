package testing_test

import (
	"testing"

	dncTesting "hostettler.dev/dnc/testing"
)

func TestBasicInstantiation(t *testing.T) {
	// Just test that we can create a test app
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Test basic getters
	if testApp.GetCharacterDir() == "" {
		t.Error("Character directory should not be empty")
	}

	if testApp.IsRunning() {
		t.Error("App should not be running before Start() is called")
	}
}

func TestTestableAppCreation(t *testing.T) {
	// Test creating the testable app model
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	// Basic checks
	if model.GetCurrentScreenType() == "" {
		t.Error("Screen type should not be empty")
	}

	t.Logf("Current screen: %s", model.GetCurrentScreenType())
}