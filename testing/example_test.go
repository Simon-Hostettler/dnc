package testing_test

import (
	"testing"
	"time"

	dncTesting "hostettler.dev/dnc/testing"
)

func TestBasicAppInstantiation(t *testing.T) {
	// Create a new test app
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create a testable app model
	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	// Create test app from model
	appWithModel, err := dncTesting.NewTestAppFromModel(model)
	if err != nil {
		t.Fatalf("Failed to create test app from model: %v", err)
	}
	defer appWithModel.Cleanup()

	// Start the test app
	if err := appWithModel.Start(); err != nil {
		t.Fatalf("Failed to start test app: %v", err)
	}

	// Test basic functionality
	if !appWithModel.IsRunning() {
		t.Error("Test app should be running")
	}

	// Check initial screen
	currentScreen := appWithModel.GetCurrentScreen()
	if currentScreen != "title" {
		t.Errorf("Expected initial screen to be 'title', got '%s'", currentScreen)
	}
}

func TestKeypressHandling(t *testing.T) {
	// Create and start test app
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	appWithModel, err := dncTesting.NewTestAppFromModel(model)
	if err != nil {
		t.Fatalf("Failed to create test app from model: %v", err)
	}
	defer appWithModel.Cleanup()

	if err := appWithModel.Start(); err != nil {
		t.Fatalf("Failed to start test app: %v", err)
	}

	// Test sending individual keys
	keys := []string{"up", "down", "left", "right", "enter", "escape"}
	for _, key := range keys {
		if err := appWithModel.SendKey(key); err != nil {
			t.Errorf("Failed to send key '%s': %v", key, err)
		}
	}

	// Test sending multiple keys
	if err := appWithModel.SendKeys("up", "down", "enter"); err != nil {
		t.Errorf("Failed to send multiple keys: %v", err)
	}

	// Test sending text
	if err := appWithModel.SendText("TestCharacter"); err != nil {
		t.Errorf("Failed to send text: %v", err)
	}
}

func TestCharacterCreation(t *testing.T) {
	// Create test app
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create a test character file
	err = testApp.CreateTestCharacter("TestHero")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Create testable app model
	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	appWithModel, err := dncTesting.NewTestAppFromModel(model)
	if err != nil {
		t.Fatalf("Failed to create test app from model: %v", err)
	}
	defer appWithModel.Cleanup()

	if err := appWithModel.Start(); err != nil {
		t.Fatalf("Failed to start test app: %v", err)
	}

	// Test that we start on title screen
	if appWithModel.GetCurrentScreen() != "title" {
		t.Error("Should start on title screen")
	}
}

func TestScreenNavigation(t *testing.T) {
	// Create test app
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	appWithModel, err := dncTesting.NewTestAppFromModel(model)
	if err != nil {
		t.Fatalf("Failed to create test app from model: %v", err)
	}
	defer appWithModel.Cleanup()

	if err := appWithModel.Start(); err != nil {
		t.Fatalf("Failed to start test app: %v", err)
	}

	// Test navigation
	initialScreen := appWithModel.GetCurrentScreen()
	t.Logf("Initial screen: %s", initialScreen)

	// Try to navigate to title screen
	if err := appWithModel.NavigateToTitleScreen(); err != nil {
		t.Errorf("Failed to navigate to title screen: %v", err)
	}

	// Verify we're on title screen
	if appWithModel.GetCurrentScreen() != "title" {
		t.Error("Should be on title screen after navigation")
	}
}

func TestOutputCapture(t *testing.T) {
	// Create test app
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	appWithModel, err := dncTesting.NewTestAppFromModel(model)
	if err != nil {
		t.Fatalf("Failed to create test app from model: %v", err)
	}
	defer appWithModel.Cleanup()

	if err := appWithModel.Start(); err != nil {
		t.Fatalf("Failed to start test app: %v", err)
	}

	// Wait a moment for rendering
	time.Sleep(200 * time.Millisecond)

	// Get output
	output := appWithModel.GetOutput()
	if output == "" {
		t.Error("Expected some output from the app")
	}

	t.Logf("App output length: %d characters", len(output))

	// Test output clearing
	appWithModel.ClearOutput()
	if appWithModel.GetOutput() != "" {
		t.Error("Output should be empty after clearing")
	}
}

func TestAppCleanup(t *testing.T) {
	// Create test app
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	characterDir := testApp.GetCharacterDir()
	
	// Cleanup should remove the temp directory
	testApp.Cleanup()

	// Verify the app is no longer running
	if testApp.IsRunning() {
		t.Error("App should not be running after cleanup")
	}

	// Note: We can't easily test if the temp directory was removed
	// because it's cleaned up in a separate goroutine
	t.Logf("Test character directory was: %s", characterDir)
}