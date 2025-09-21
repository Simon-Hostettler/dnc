# End-to-End Testing Framework for D&C App

This testing framework provides a comprehensive solution for testing the D&D character (D&C) application's behavior through simulated user interactions. You can instantiate the app and send keypresses to test its functionality without running the full TUI interface.

## Overview

The testing framework consists of several key components:

- **TestApp**: The main testing harness that manages app lifecycle and provides testing utilities
- **TestableDnCApp**: A testable version of the main application that implements the required interfaces
- **TestableApp interface**: Defines the contract for testable applications
- **Helper functions**: Various utilities for common testing scenarios

## Getting Started

### Basic Usage

```go
package main

import (
    "testing"
    "hostettler.dev/dnc/testing"
)

func TestBasicAppFunctionality(t *testing.T) {
    // Create a new test app with temporary directory
    testApp, err := testing.NewTestApp()
    if err != nil {
        t.Fatalf("Failed to create test app: %v", err)
    }
    defer testApp.Cleanup() // Always cleanup resources

    // Create the testable app model
    model, err := testing.NewTestableApp(testApp.GetCharacterDir())
    if err != nil {
        t.Fatalf("Failed to create testable app: %v", err)
    }

    // Create test app from model
    appWithModel, err := testing.NewTestAppFromModel(model)
    if err != nil {
        t.Fatalf("Failed to create test app from model: %v", err)
    }
    defer appWithModel.Cleanup()

    // Start the test app
    if err := appWithModel.Start(); err != nil {
        t.Fatalf("Failed to start test app: %v", err)
    }

    // Now you can test the app behavior
    currentScreen := appWithModel.GetCurrentScreen()
    // ... perform tests
}
```

## Key Features

### 1. Keypress Simulation

Send individual keys or sequences of keys to the application:

```go
// Send individual keys
err := testApp.SendKey("up")
err = testApp.SendKey("down")
err = testApp.SendKey("enter")

// Send multiple keys in sequence
err = testApp.SendKeys("up", "up", "down", "enter")

// Send text as individual character keypresses
err = testApp.SendText("MyCharacterName")
```

#### Supported Key Types

- **Navigation**: "up", "down", "left", "right"
- **Actions**: "enter", "escape", "space", "tab"
- **Editing**: "backspace", "delete"
- **Control**: "ctrl+c"
- **Text**: Any single character or string

### 2. Screen Detection and Navigation

The framework can detect which screen is currently active and provides navigation helpers:

```go
// Get current screen
currentScreen := testApp.GetCurrentScreen() // Returns: "title", "score", "editor", or "unknown"

// Navigate to specific screens
err := testApp.NavigateToTitleScreen()

// Wait for specific screen to appear
err := testApp.WaitForScreen("score", 5*time.Second)
```

### 3. Output Capture and Validation

Capture and analyze the app's output:

```go
// Get current output
output := testApp.GetOutput()

// Wait for specific text to appear
err := testApp.WaitForOutput("Create new Character", 2*time.Second)

// Clear output buffer
testApp.ClearOutput()
```

### 4. Character Management

Create test characters and manage character data:

```go
// Create a test character file
err := testApp.CreateTestCharacter("TestHero")

// Get current character from app
character := testApp.GetCharacter()
```

### 5. App State Inspection

Access the underlying app state for validation:

```go
// Check if app is running
isRunning := testApp.IsRunning()

// Get character directory
charDir := testApp.GetCharacterDir()

// Set terminal size
err := testApp.SetSize(120, 40)
```

## Testing Patterns

### Testing Navigation

```go
func TestNavigation(t *testing.T) {
    testApp := setupTestApp(t) // helper function
    defer testApp.Cleanup()

    // Test moving between screens
    err := testApp.SendKeys("right", "enter")
    if err != nil {
        t.Errorf("Navigation failed: %v", err)
    }

    // Verify we're on the expected screen
    if testApp.GetCurrentScreen() != "score" {
        t.Error("Should be on score screen")
    }
}
```

### Testing Character Creation

```go
func TestCharacterCreation(t *testing.T) {
    testApp := setupTestApp(t)
    defer testApp.Cleanup()

    // Create character through UI
    testApp.SendText("NewHero")
    testApp.SendKey("enter")

    // Verify character was created
    character := testApp.GetCharacter()
    if character == nil || character.Name != "NewHero" {
        t.Error("Character creation failed")
    }
}
```

### Testing Input Validation

```go
func TestInputValidation(t *testing.T) {
    testApp := setupTestApp(t)
    defer testApp.Cleanup()

    // Test invalid input
    testApp.SendText("InvalidCharacterName123!")
    testApp.SendKey("enter")

    // Check for error message
    err := testApp.WaitForOutput("Invalid character name", 1*time.Second)
    if err != nil {
        t.Error("Should show validation error")
    }
}
```

## Advanced Usage

### Custom Test Scenarios

You can create custom helper functions for complex test scenarios:

```go
func createCharacterAndNavigateToStats(testApp *testing.TestApp, characterName string) error {
    // Create character
    if err := testApp.CreateTestCharacter(characterName); err != nil {
        return err
    }

    // Navigate to character selection
    if err := testApp.NavigateToTitleScreen(); err != nil {
        return err
    }

    // Select character
    if err := testApp.SendKey("enter"); err != nil {
        return err
    }

    // Wait for stats screen
    return testApp.WaitForScreen("score", 2*time.Second)
}
```

### Parallel Testing

The framework supports parallel testing since each test gets its own temporary directory:

```go
func TestParallelNavigation(t *testing.T) {
    t.Parallel()
    testApp := setupTestApp(t)
    defer testApp.Cleanup()
    // ... test logic
}
```

## Best Practices

1. **Always use defer for cleanup**: Ensure `testApp.Cleanup()` is called to remove temporary files
2. **Use timeouts**: When waiting for output or screen changes, always specify reasonable timeouts
3. **Test realistic scenarios**: Simulate actual user interactions rather than testing individual functions
4. **Validate state changes**: After sending keys, verify the app state changed as expected
5. **Handle errors**: Always check return values from testing framework methods

## Error Handling

The framework provides descriptive error messages for common issues:

- App not started: Call `Start()` before sending keys
- Invalid keys: The framework supports a predefined set of key types
- Timeouts: Operations that wait for changes include timeout handling
- Resource cleanup: Automatic cleanup of temporary directories and processes

## Limitations

1. **Output parsing**: Screen detection relies on text analysis which may be fragile
2. **Timing**: Some operations require delays to allow the app to process changes
3. **Platform dependencies**: Testing may behave differently on different operating systems
4. **Complex UI interactions**: Some advanced UI patterns may be difficult to test

## Running Tests

```bash
# Run all tests in the testing package
go test ./testing

# Run with verbose output
go test -v ./testing

# Run a specific test
go test -v ./testing -run TestBasicAppInstantiation

# Run tests in parallel
go test -parallel 4 ./testing
```

## Troubleshooting

### Common Issues

1. **Tests hanging**: Check for missing `Cleanup()` calls or infinite loops in app logic
2. **Screen detection failing**: Verify output contains expected text patterns
3. **Keypress not working**: Ensure app is started and key type is supported
4. **Temporary directory issues**: Make sure tests have write permissions

### Debug Output

Use the testing framework's output capture to debug issues:

```go
output := testApp.GetOutput()
t.Logf("App output: %s", output)
```

This comprehensive testing framework enables thorough end-to-end testing of the D&C application's user interface and behavior through simulated user interactions.