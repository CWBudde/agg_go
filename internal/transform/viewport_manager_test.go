package transform

import (
	"testing"
)

func TestNewViewportManager(t *testing.T) {
	vm := NewViewportManager()

	if vm.GetCurrentName() != "root" {
		t.Error("New viewport manager should start with root viewport as current")
	}

	if vm.GetCurrent() == nil {
		t.Error("Root viewport should not be nil")
	}

	names := vm.ListViewports()
	if len(names) != 1 || names[0] != "root" {
		t.Errorf("New manager should have only root viewport, got %v", names)
	}
}

func TestCreateViewport(t *testing.T) {
	vm := NewViewportManager()

	err := vm.CreateViewport("test", 0.0, 0.0, 100.0, 100.0, 0.0, 0.0, 800.0, 600.0)
	if err != nil {
		t.Fatalf("CreateViewport failed: %v", err)
	}

	if !vm.HasViewport("test") {
		t.Error("Created viewport should exist")
	}

	viewport := vm.GetViewport("test")
	if viewport == nil {
		t.Error("GetViewport should return the created viewport")
	}

	wx1, wy1, wx2, wy2 := viewport.GetWorldViewport()
	if wx1 != 0.0 || wy1 != 0.0 || wx2 != 100.0 || wy2 != 100.0 {
		t.Errorf("Viewport world bounds incorrect, got (%g,%g)-(%g,%g)", wx1, wy1, wx2, wy2)
	}
}

func TestCreateViewportEmptyName(t *testing.T) {
	vm := NewViewportManager()

	err := vm.CreateViewport("", 0.0, 0.0, 100.0, 100.0, 0.0, 0.0, 800.0, 600.0)
	if err == nil {
		t.Error("CreateViewport with empty name should fail")
	}
}

func TestSwitchTo(t *testing.T) {
	vm := NewViewportManager()

	// Create a viewport
	vm.CreateViewport("test", 0.0, 0.0, 100.0, 100.0, 0.0, 0.0, 800.0, 600.0)

	// Switch to it
	err := vm.SwitchTo("test")
	if err != nil {
		t.Fatalf("SwitchTo failed: %v", err)
	}

	if vm.GetCurrentName() != "test" {
		t.Error("Current viewport name should be 'test'")
	}

	// Switch back to root
	err = vm.SwitchTo("root")
	if err != nil {
		t.Fatalf("SwitchTo root failed: %v", err)
	}

	if vm.GetCurrentName() != "root" {
		t.Error("Current viewport name should be 'root'")
	}
}

func TestSwitchToNonexistent(t *testing.T) {
	vm := NewViewportManager()

	err := vm.SwitchTo("nonexistent")
	if err == nil {
		t.Error("SwitchTo nonexistent viewport should fail")
	}

	// Current should remain unchanged
	if vm.GetCurrentName() != "root" {
		t.Error("Current viewport should remain 'root' after failed switch")
	}
}

func TestRemoveViewport(t *testing.T) {
	vm := NewViewportManager()

	// Create and switch to a viewport
	vm.CreateViewport("test", 0.0, 0.0, 100.0, 100.0, 0.0, 0.0, 800.0, 600.0)
	vm.SwitchTo("test")

	// Remove it
	err := vm.RemoveViewport("test")
	if err != nil {
		t.Fatalf("RemoveViewport failed: %v", err)
	}

	if vm.HasViewport("test") {
		t.Error("Removed viewport should not exist")
	}

	// Should have switched back to root
	if vm.GetCurrentName() != "root" {
		t.Error("Should have switched to root after removing current viewport")
	}
}

func TestRemoveRootViewport(t *testing.T) {
	vm := NewViewportManager()

	err := vm.RemoveViewport("root")
	if err == nil {
		t.Error("Removing root viewport should fail")
	}
}

func TestCopyViewport(t *testing.T) {
	vm := NewViewportManager()

	// Create a viewport with specific settings
	vm.CreateViewport("source", 10.0, 20.0, 110.0, 120.0, 0.0, 0.0, 800.0, 600.0)
	source := vm.GetViewport("source")
	source.PreserveAspectRatio(0.25, 0.75, AspectRatioMeet)

	// Copy it
	err := vm.CopyViewport("source", "copy")
	if err != nil {
		t.Fatalf("CopyViewport failed: %v", err)
	}

	copy := vm.GetViewport("copy")
	if copy == nil {
		t.Fatal("Copied viewport should exist")
	}

	// Check that settings were copied
	wx1, wy1, wx2, wy2 := copy.GetWorldViewport()
	if wx1 != 10.0 || wy1 != 20.0 || wx2 != 110.0 || wy2 != 120.0 {
		t.Errorf("Copied viewport world bounds incorrect, got (%g,%g)-(%g,%g)", wx1, wy1, wx2, wy2)
	}

	if copy.AlignX() != 0.25 || copy.AlignY() != 0.75 || copy.AspectRatio() != AspectRatioMeet {
		t.Error("Copied viewport aspect ratio settings incorrect")
	}
}

func TestClear(t *testing.T) {
	vm := NewViewportManager()

	// Create some viewports
	vm.CreateViewport("test1", 0.0, 0.0, 100.0, 100.0, 0.0, 0.0, 800.0, 600.0)
	vm.CreateViewport("test2", 0.0, 0.0, 100.0, 100.0, 0.0, 0.0, 800.0, 600.0)
	vm.SwitchTo("test1")

	vm.Clear()

	names := vm.ListViewports()
	if len(names) != 1 || names[0] != "root" {
		t.Errorf("After clear, should only have root viewport, got %v", names)
	}

	if vm.GetCurrentName() != "root" {
		t.Error("After clear, current should be root")
	}
}

func TestViewportStack(t *testing.T) {
	vm := NewViewportManager()
	stack := NewViewportStack(vm)

	// Modify the root viewport
	root := vm.GetCurrent()
	root.WorldViewport(10.0, 20.0, 110.0, 120.0)

	// Push current state
	stack.Push()

	if stack.Depth() != 1 {
		t.Error("Stack depth should be 1 after push")
	}

	// Modify viewport
	root.WorldViewport(50.0, 60.0, 150.0, 160.0)

	// Pop should restore original state
	success := stack.Pop()
	if !success {
		t.Error("Pop should succeed")
	}

	if stack.Depth() != 0 {
		t.Error("Stack depth should be 0 after pop")
	}

	// Check that original state was restored
	wx1, wy1, wx2, wy2 := root.GetWorldViewport()
	if wx1 != 10.0 || wy1 != 20.0 || wx2 != 110.0 || wy2 != 120.0 {
		t.Errorf("Pop should restore original bounds, got (%g,%g)-(%g,%g)", wx1, wy1, wx2, wy2)
	}
}

func TestViewportStackEmpty(t *testing.T) {
	vm := NewViewportManager()
	stack := NewViewportStack(vm)

	// Pop from empty stack should fail
	success := stack.Pop()
	if success {
		t.Error("Pop from empty stack should return false")
	}
}

func TestViewportStackClear(t *testing.T) {
	vm := NewViewportManager()
	stack := NewViewportStack(vm)

	// Push some states
	stack.Push()
	stack.Push()
	stack.Push()

	if stack.Depth() != 3 {
		t.Error("Stack depth should be 3")
	}

	stack.Clear()

	if stack.Depth() != 0 {
		t.Error("Stack depth should be 0 after clear")
	}
}
