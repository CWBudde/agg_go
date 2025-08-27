// Package transform provides viewport management functionality for AGG.
// This implements multi-viewport support for complex applications.
package transform

import (
	"fmt"
)

// ViewportManager manages multiple named viewports and provides
// functionality for viewport hierarchies and switching.
type ViewportManager struct {
	viewports map[string]*TransViewport
	current   string
	root      *TransViewport // Default viewport
}

// NewViewportManager creates a new viewport manager with a default root viewport.
func NewViewportManager() *ViewportManager {
	root := NewTransViewport()
	return &ViewportManager{
		viewports: make(map[string]*TransViewport),
		current:   "root",
		root:      root,
	}
}

// CreateViewport creates a new named viewport with the given parameters.
// If a viewport with the same name already exists, it will be replaced.
func (vm *ViewportManager) CreateViewport(name string, worldX1, worldY1, worldX2, worldY2, deviceX1, deviceY1, deviceX2, deviceY2 float64) error {
	if name == "" {
		return fmt.Errorf("viewport name cannot be empty")
	}

	viewport := NewTransViewport()
	viewport.WorldViewport(worldX1, worldY1, worldX2, worldY2)
	viewport.DeviceViewport(deviceX1, deviceY1, deviceX2, deviceY2)

	vm.viewports[name] = viewport
	return nil
}

// AddViewport adds an existing viewport with the given name.
// If a viewport with the same name already exists, it will be replaced.
func (vm *ViewportManager) AddViewport(name string, viewport *TransViewport) error {
	if name == "" {
		return fmt.Errorf("viewport name cannot be empty")
	}
	if viewport == nil {
		return fmt.Errorf("viewport cannot be nil")
	}

	vm.viewports[name] = viewport
	return nil
}

// RemoveViewport removes the named viewport.
// The root viewport cannot be removed.
// If the removed viewport was the current one, switches to root.
func (vm *ViewportManager) RemoveViewport(name string) error {
	if name == "root" {
		return fmt.Errorf("cannot remove root viewport")
	}

	if _, exists := vm.viewports[name]; !exists {
		return fmt.Errorf("viewport '%s' does not exist", name)
	}

	delete(vm.viewports, name)

	// If we just removed the current viewport, switch to root
	if vm.current == name {
		vm.current = "root"
	}

	return nil
}

// SwitchTo switches to the named viewport, making it the current active viewport.
func (vm *ViewportManager) SwitchTo(name string) error {
	// Check if it's the root viewport
	if name == "root" {
		vm.current = "root"
		return nil
	}

	// Check if the named viewport exists
	if _, exists := vm.viewports[name]; !exists {
		return fmt.Errorf("viewport '%s' does not exist", name)
	}

	vm.current = name
	return nil
}

// GetCurrent returns the currently active viewport.
func (vm *ViewportManager) GetCurrent() *TransViewport {
	if vm.current == "root" {
		return vm.root
	}

	if viewport, exists := vm.viewports[vm.current]; exists {
		return viewport
	}

	// Fallback to root if current viewport was somehow deleted
	vm.current = "root"
	return vm.root
}

// GetViewport returns the viewport with the given name.
// Returns nil if the viewport doesn't exist.
func (vm *ViewportManager) GetViewport(name string) *TransViewport {
	if name == "root" {
		return vm.root
	}

	return vm.viewports[name]
}

// GetCurrentName returns the name of the currently active viewport.
func (vm *ViewportManager) GetCurrentName() string {
	return vm.current
}

// ListViewports returns a slice of all viewport names.
func (vm *ViewportManager) ListViewports() []string {
	names := make([]string, 0, len(vm.viewports)+1)
	names = append(names, "root")

	for name := range vm.viewports {
		names = append(names, name)
	}

	return names
}

// HasViewport returns true if a viewport with the given name exists.
func (vm *ViewportManager) HasViewport(name string) bool {
	if name == "root" {
		return true
	}

	_, exists := vm.viewports[name]
	return exists
}

// CopyViewport creates a copy of an existing viewport with a new name.
func (vm *ViewportManager) CopyViewport(sourceName, destName string) error {
	if destName == "" {
		return fmt.Errorf("destination viewport name cannot be empty")
	}

	source := vm.GetViewport(sourceName)
	if source == nil {
		return fmt.Errorf("source viewport '%s' does not exist", sourceName)
	}

	// Create a new viewport with the same settings
	copy := NewTransViewport()
	wx1, wy1, wx2, wy2 := source.GetWorldViewport()
	dx1, dy1, dx2, dy2 := source.GetDeviceViewport()

	copy.WorldViewport(wx1, wy1, wx2, wy2)
	copy.DeviceViewport(dx1, dy1, dx2, dy2)
	copy.PreserveAspectRatio(source.AlignX(), source.AlignY(), source.AspectRatio())

	vm.viewports[destName] = copy
	return nil
}

// Clear removes all viewports except the root viewport and switches to root.
func (vm *ViewportManager) Clear() {
	vm.viewports = make(map[string]*TransViewport)
	vm.current = "root"
}

// ViewportStack manages a stack of viewport states for push/pop operations.
type ViewportStack struct {
	manager *ViewportManager
	stack   []ViewportState
}

// ViewportState represents a saved viewport state.
type ViewportState struct {
	name     string
	viewport *TransViewport
}

// NewViewportStack creates a new viewport stack for the given manager.
func NewViewportStack(manager *ViewportManager) *ViewportStack {
	return &ViewportStack{
		manager: manager,
		stack:   make([]ViewportState, 0),
	}
}

// Push saves the current viewport state onto the stack.
func (vs *ViewportStack) Push() {
	current := vs.manager.GetCurrent()
	currentName := vs.manager.GetCurrentName()

	// Create a copy of the current viewport
	copy := NewTransViewport()
	wx1, wy1, wx2, wy2 := current.GetWorldViewport()
	dx1, dy1, dx2, dy2 := current.GetDeviceViewport()

	copy.WorldViewport(wx1, wy1, wx2, wy2)
	copy.DeviceViewport(dx1, dy1, dx2, dy2)
	copy.PreserveAspectRatio(current.AlignX(), current.AlignY(), current.AspectRatio())

	state := ViewportState{
		name:     currentName,
		viewport: copy,
	}

	vs.stack = append(vs.stack, state)
}

// Pop restores the most recently saved viewport state from the stack.
// Returns false if the stack is empty.
func (vs *ViewportStack) Pop() bool {
	if len(vs.stack) == 0 {
		return false
	}

	// Get the last state
	state := vs.stack[len(vs.stack)-1]
	vs.stack = vs.stack[:len(vs.stack)-1]

	// Restore the viewport state
	current := vs.manager.GetCurrent()
	wx1, wy1, wx2, wy2 := state.viewport.GetWorldViewport()
	dx1, dy1, dx2, dy2 := state.viewport.GetDeviceViewport()

	current.WorldViewport(wx1, wy1, wx2, wy2)
	current.DeviceViewport(dx1, dy1, dx2, dy2)
	current.PreserveAspectRatio(state.viewport.AlignX(), state.viewport.AlignY(), state.viewport.AspectRatio())

	return true
}

// Depth returns the current depth of the viewport stack.
func (vs *ViewportStack) Depth() int {
	return len(vs.stack)
}

// Clear empties the viewport stack.
func (vs *ViewportStack) Clear() {
	vs.stack = vs.stack[:0]
}
