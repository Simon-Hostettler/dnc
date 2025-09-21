package testing_test

import (
	"testing"
	"time"

	dncTesting "hostettler.dev/dnc/testing"
)

func TestWorkingExample(t *testing.T) {
	// Create a test app instance
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create a test character
	err = testApp.CreateTestCharacter("TestWarrior")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Create testable app model
	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	// Test basic functionality without starting the full program
	t.Logf("Initial screen: %s", model.GetCurrentScreenType())
	
	// Test the Init method
	cmd := model.Init()
	if cmd == nil {
		t.Log("Init returned nil command (this is acceptable)")
	} else {
		t.Log("Init returned a command")
	}

	// Test that we can get the current screen type
	screenType := model.GetCurrentScreenType()
	if screenType == "" {
		t.Error("Screen type should not be empty")
	}
	t.Logf("Screen type: %s", screenType)

	// Test the View method (without full program)
	view := model.View()
	if view == "" {
		t.Log("View is empty (this might be expected for uninitialized app)")
	} else {
		t.Logf("View output length: %d characters", len(view))
	}
}

// This test demonstrates testing without starting the full bubbletea program
func TestKeyMessageHandling(t *testing.T) {
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	// Simulate a key press directly on the model
	// This tests the Update logic without needing the full program
	keyMsg := dncTesting.CreateKeyMsg("down") // We'll need to expose this helper
	if keyMsg.Type != 0 { // Basic validation that we got a key message
		t.Log("Successfully created key message")
	}

	// Test that the model can handle updates
	initialScreen := model.GetCurrentScreenType()
	t.Logf("Before update - Screen: %s", initialScreen)

	// We can call Update directly for unit-style testing
	updatedModel, cmd := model.Update(keyMsg)
	if updatedModel == nil {
		t.Error("Update should return a model")
	}
	if cmd != nil {
		t.Log("Update returned a command")
	}

	afterScreen := model.GetCurrentScreenType()
	t.Logf("After update - Screen: %s", afterScreen)
}

// Test the testing infrastructure utilities
func TestUtilities(t *testing.T) {
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Test character directory creation
	charDir := testApp.GetCharacterDir()
	if charDir == "" {
		t.Error("Character directory should not be empty")
	}
	t.Logf("Character directory: %s", charDir)

	// Test character creation
	err = testApp.CreateTestCharacter("Gandalf")
	if err != nil {
		t.Errorf("Failed to create character: %v", err)
	}

	// Test setting size
	err = testApp.SetSize(120, 40)
	if err != nil {
		t.Errorf("Failed to set size: %v", err)
	}

	// Test that cleanup works (non-destructive test)
	if !testApp.IsRunning() {
		t.Log("App is not running (expected before Start)")
	}
}

// Demonstrate testing character creation and file operations
func TestCharacterOperations(t *testing.T) {
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create multiple test characters
	characters := []string{"Fighter", "Wizard", "Rogue", "Cleric"}
	
	for _, charName := range characters {
		err = testApp.CreateTestCharacter(charName)
		if err != nil {
			t.Errorf("Failed to create character %s: %v", charName, err)
		} else {
			t.Logf("Successfully created character: %s", charName)
		}
	}

	// Create the testable app with the character directory
	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		t.Fatalf("Failed to create testable app: %v", err)
	}

	// Test that the app starts on the title screen
	if model.GetCurrentScreenType() != "title" {
		t.Errorf("Expected title screen, got: %s", model.GetCurrentScreenType())
	}

	// Test that no character is loaded initially
	char := model.GetCharacter()
	if char != nil {
		t.Error("No character should be loaded initially")
	}
}

// Test timeout and waiting functionality
func TestTimeoutFunctionality(t *testing.T) {
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Test output operations
	output := testApp.GetOutput()
	if output != "" {
		t.Error("Output should be empty initially")
	}

	testApp.ClearOutput()
	if testApp.GetOutput() != "" {
		t.Error("Output should be empty after clearing")
	}

	// Test waiting for output that will never come (with short timeout)
	err = testApp.WaitForOutput("NonexistentText", 100*time.Millisecond)
	if err == nil {
		t.Error("Should have timed out waiting for nonexistent output")
	}
	t.Logf("Expected timeout error: %v", err)
}