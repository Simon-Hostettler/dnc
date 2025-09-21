// Package testing provides a complete end-to-end testing framework for the D&C application.
// This package allows you to instantiate the app and send keypresses to test its behavior.
package testing

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
)

// TestableApp interface defines what our testable app needs to implement
type TestableApp interface {
	tea.Model
	GetCurrentScreenType() string
	GetCharacter() *models.Character
}

// TestApp provides a testable version of the D&C application
type TestApp struct {
	program        *tea.Program
	model          TestableApp
	output         *bytes.Buffer
	width          int
	height         int
	characterDir   string
	tempDirCleanup func()
	running        bool
}

// NewTestApp creates a new test app with a temporary character directory
func NewTestApp() (*TestApp, error) {
	// Create temporary directory for test characters
	tempDir, err := os.MkdirTemp("", "dnc-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	characterDir := filepath.Join(tempDir, "characters")
	if err := os.MkdirAll(characterDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create character dir: %w", err)
	}

	output := &bytes.Buffer{}
	
	testApp := &TestApp{
		output:       output,
		width:        80,
		height:       24,
		characterDir: characterDir,
		tempDirCleanup: func() {
			os.RemoveAll(tempDir)
		},
	}

	return testApp, nil
}

// NewTestAppFromModel creates a test app from an existing model that implements TestableApp
func NewTestAppFromModel(model TestableApp) (*TestApp, error) {
	// Create temporary directory for test characters
	tempDir, err := os.MkdirTemp("", "dnc-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	characterDir := filepath.Join(tempDir, "characters")
	if err := os.MkdirAll(characterDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create character dir: %w", err)
	}

	output := &bytes.Buffer{}
	
	testApp := &TestApp{
		model:        model,
		output:       output,
		width:        80,
		height:       24,
		characterDir: characterDir,
		tempDirCleanup: func() {
			os.RemoveAll(tempDir)
		},
	}

	return testApp, nil
}

// Start initializes the test app and prepares it for testing
func (ta *TestApp) Start() error {
	if ta.model == nil {
		return fmt.Errorf("no model provided")
	}

	// Create a program with custom output for testing
	ta.program = tea.NewProgram(
		ta.model,
		tea.WithOutput(ta.output),
		tea.WithoutSignals(),
		tea.WithoutCatchPanics(),
	)

	// Send initial window size
	ta.program.Send(tea.WindowSizeMsg{Width: ta.width, Height: ta.height})

	ta.running = true
	
	// Give it a moment to initialize
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// SendKey sends a key press to the application
func (ta *TestApp) SendKey(key string) error {
	if !ta.running {
		return fmt.Errorf("test app not started")
	}

	if ta.program == nil {
		return fmt.Errorf("program not initialized")
	}

	keyMsg := ta.createKeyMsg(key)
	ta.program.Send(keyMsg)
	
	// Give the program time to process the key
	time.Sleep(50 * time.Millisecond)
	
	return nil
}

// createKeyMsg creates appropriate tea.KeyMsg for different key types
func (ta *TestApp) createKeyMsg(key string) tea.KeyMsg {
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "escape":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "delete":
		return tea.KeyMsg{Type: tea.KeyDelete}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		// For single characters
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(key[0])}}
		}
		// For multi-character strings, send as runes
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}

// SendKeys sends multiple key presses in sequence
func (ta *TestApp) SendKeys(keys ...string) error {
	for _, key := range keys {
		if err := ta.SendKey(key); err != nil {
			return err
		}
		// Small delay between key presses to simulate realistic usage
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// SendText sends text as individual character key presses
func (ta *TestApp) SendText(text string) error {
	for _, char := range text {
		if err := ta.SendKey(string(char)); err != nil {
			return err
		}
	}
	return nil
}

// GetOutput returns the current output buffer content
func (ta *TestApp) GetOutput() string {
	if ta.output == nil {
		return ""
	}
	return ta.output.String()
}

// ClearOutput clears the output buffer
func (ta *TestApp) ClearOutput() {
	if ta.output != nil {
		ta.output.Reset()
	}
}

// GetCurrentScreen attempts to determine which screen is currently active
func (ta *TestApp) GetCurrentScreen() string {
	if ta.model != nil {
		return ta.model.GetCurrentScreenType()
	}
	
	// Fallback to output parsing
	output := ta.GetOutput()
	
	if strings.Contains(output, "Create new Character") {
		return "title"
	}
	if strings.Contains(output, "Stats") {
		return "score"
	}
	if strings.Contains(output, "Save") {
		return "editor"
	}
	
	return "unknown"
}

// WaitForOutput waits for specific text to appear in the output
func (ta *TestApp) WaitForOutput(expectedText string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if strings.Contains(ta.GetOutput(), expectedText) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for output: %s", expectedText)
}

// WaitForScreen waits for a specific screen to become active
func (ta *TestApp) WaitForScreen(screenName string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if ta.GetCurrentScreen() == screenName {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for screen: %s", screenName)
}

// CreateTestCharacter creates a test character file in the test directory
func (ta *TestApp) CreateTestCharacter(name string) error {
	char, err := models.NewCharacter(name)
	if err != nil {
		return err
	}
	
	// Set some default values for testing
	char.Race = "Human"
	char.ClassLevels = "Fighter 1"
	char.Abilities.Strength = 15
	char.Abilities.Dexterity = 14
	char.Abilities.Constitution = 13
	
	// Save to test directory
	filename := fmt.Sprintf("%s.json", strings.ToLower(name))
	char.SaveFile = filepath.Join(ta.characterDir, filename)
	
	return char.SaveToFile()
}

// GetCharacter returns the current character from the model
func (ta *TestApp) GetCharacter() *models.Character {
	if ta.model != nil {
		return ta.model.GetCharacter()
	}
	return nil
}

// Cleanup cleans up temporary files and resources
func (ta *TestApp) Cleanup() {
	if ta.program != nil {
		ta.program.Kill()
	}
	if ta.tempDirCleanup != nil {
		ta.tempDirCleanup()
	}
	ta.running = false
}

// Stop gracefully stops the test app
func (ta *TestApp) Stop() {
	if ta.program != nil {
		ta.program.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
		ta.program.Wait()
	}
	ta.running = false
}

// GetCharacterDir returns the temporary character directory path
func (ta *TestApp) GetCharacterDir() string {
	return ta.characterDir
}

// IsRunning returns whether the test app is currently running
func (ta *TestApp) IsRunning() bool {
	return ta.running
}

// SetSize sets the terminal size for the test app
func (ta *TestApp) SetSize(width, height int) error {
	ta.width = width
	ta.height = height
	
	if ta.program != nil {
		ta.program.Send(tea.WindowSizeMsg{Width: width, Height: height})
	}
	
	return nil
}

// Helper methods for common testing scenarios

// NavigateToTitleScreen attempts to navigate to the title screen
func (ta *TestApp) NavigateToTitleScreen() error {
	// Try pressing escape a few times
	for i := 0; i < 5; i++ {
		currentScreen := ta.GetCurrentScreen()
		if currentScreen == "title" {
			return nil
		}
		ta.SendKey("escape")
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("could not navigate to title screen")
}

// SelectCharacter attempts to select a character from the character list
func (ta *TestApp) SelectCharacter(characterName string) error {
	if ta.GetCurrentScreen() != "title" {
		if err := ta.NavigateToTitleScreen(); err != nil {
			return err
		}
	}
	
	// The implementation would depend on how character selection works in the UI
	// For now, just send enter to select the current character
	return ta.SendKey("enter")
}

// CreateNewCharacter creates a new character through the UI
func (ta *TestApp) CreateNewCharacter(name string) error {
	if ta.GetCurrentScreen() != "title" {
		if err := ta.NavigateToTitleScreen(); err != nil {
			return err
		}
	}
	
	// This would need to be implemented based on the actual UI flow
	// For now, assume there's a "new character" option
	return ta.SendText(name)
}