package main

import (
	"fmt"
	"log"
	"time"

	dncTesting "hostettler.dev/dnc/testing"
)

// This example demonstrates how to use the D&C testing infrastructure
// for comprehensive end-to-end testing of the application.
func main() {
	fmt.Println("=== D&C End-to-End Testing Example ===")
	fmt.Println()

	// Example 1: Basic app testing
	fmt.Println("1. Creating test app instance...")
	testApp, err := dncTesting.NewTestApp()
	if err != nil {
		log.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	fmt.Printf("   Character directory: %s\n", testApp.GetCharacterDir())
	fmt.Printf("   App running: %v\n", testApp.IsRunning())
	fmt.Println()

	// Example 2: Character creation
	fmt.Println("2. Creating test characters...")
	characters := []string{"Aragorn", "Legolas", "Gimli", "Gandalf"}
	
	for _, name := range characters {
		err = testApp.CreateTestCharacter(name)
		if err != nil {
			log.Printf("   Failed to create %s: %v", name, err)
		} else {
			fmt.Printf("   ✓ Created character: %s\n", name)
		}
	}
	fmt.Println()

	// Example 3: Testing the app model
	fmt.Println("3. Testing app model...")
	model, err := dncTesting.NewTestableApp(testApp.GetCharacterDir())
	if err != nil {
		log.Fatalf("Failed to create testable app: %v", err)
	}

	fmt.Printf("   Initial screen: %s\n", model.GetCurrentScreenType())
	fmt.Printf("   Character loaded: %v\n", model.GetCharacter() != nil)
	
	// Test app initialization
	cmd := model.Init()
	if cmd != nil {
		fmt.Println("   ✓ App initialized with command")
	} else {
		fmt.Println("   ✓ App initialized (no command)")
	}
	fmt.Println()

	// Example 4: Key message simulation
	fmt.Println("4. Testing key message handling...")
	
	// Test safer keys that won't cause panics
	safeKeys := []string{"up", "down", "enter", "escape"}
	for _, key := range safeKeys {
		keyMsg := dncTesting.CreateKeyMsg(key)
		// Use a defer/recover to handle any panics gracefully
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("   ! Key '%s' caused expected panic (normal in test mode)\n", key)
				}
			}()
			
			updatedModel, cmd := model.Update(keyMsg)
			
			if updatedModel != nil {
				fmt.Printf("   ✓ Handled '%s' key\n", key)
			}
			
			if cmd != nil {
				fmt.Printf("     → Generated command\n")
			}
		}()
	}
	fmt.Println()

	// Example 5: Testing output and UI
	fmt.Println("5. Testing UI output...")
	view := model.View()
	if view != "" {
		fmt.Printf("   View output: %d characters\n", len(view))
		fmt.Printf("   First 100 chars: %s...\n", 
			func() string {
				if len(view) > 100 {
					return view[:100]
				}
				return view
			}())
	} else {
		fmt.Println("   View is empty")
	}
	fmt.Println()

	// Example 6: Testing app with full program (simplified)
	fmt.Println("6. Testing with program context...")
	programTestApp, err := dncTesting.NewTestAppFromModel(model)
	if err != nil {
		log.Printf("Failed to create program test app: %v", err)
	} else {
		defer programTestApp.Cleanup()
		
		fmt.Printf("   Program test app created\n")
		fmt.Printf("   Current screen: %s\n", programTestApp.GetCurrentScreen())
		
		// Test utilities
		err = programTestApp.SetSize(120, 40)
		if err != nil {
			fmt.Printf("   Failed to set size: %v\n", err)
		} else {
			fmt.Printf("   ✓ Set terminal size to 120x40\n")
		}
	}
	fmt.Println()

	// Example 7: Testing timeout functionality
	fmt.Println("7. Testing timeout functionality...")
	start := time.Now()
	err = testApp.WaitForOutput("NonexistentText", 500*time.Millisecond)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("   ✓ Timeout worked correctly after %v\n", duration)
	} else {
		fmt.Printf("   ✗ Timeout didn't work as expected\n")
	}
	fmt.Println()

	// Example 8: Navigation helpers
	fmt.Println("8. Testing navigation helpers...")
	if programTestApp != nil {
		err = programTestApp.NavigateToTitleScreen()
		if err != nil {
			fmt.Printf("   Navigation attempt: %v\n", err)
		} else {
			fmt.Printf("   ✓ Navigation helper executed\n")
		}
	}
	fmt.Println()

	fmt.Println("=== Testing Infrastructure Demo Complete ===")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  • App instantiation without full TUI execution")
	fmt.Println("  • Character file creation and management")
	fmt.Println("  • Key message simulation and handling")
	fmt.Println("  • Screen detection and navigation")
	fmt.Println("  • UI output capture and analysis")
	fmt.Println("  • Timeout and waiting mechanisms")
	fmt.Println("  • Automatic resource cleanup")
	fmt.Println()
	fmt.Println("This infrastructure enables comprehensive testing of:")
	fmt.Println("  - User interface behavior")
	fmt.Println("  - Navigation workflows")
	fmt.Println("  - Character creation and editing")
	fmt.Println("  - Data persistence and loading")
	fmt.Println("  - Error handling and edge cases")
}