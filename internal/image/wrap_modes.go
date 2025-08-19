// Package image provides image accessor implementations for AGG.
// This package handles various pixel access patterns and boundary conditions.
package image

import "agg_go/internal/basics"

// WrapModeRepeat implements texture wrapping with repeat (tile) mode.
// Uses modulo arithmetic to wrap coordinates.
type WrapModeRepeat struct {
	size  basics.Int32u
	add   basics.Int32u
	value basics.Int32u
}

// NewWrapModeRepeat creates a new repeat wrap mode for the given size.
func NewWrapModeRepeat(size basics.Int32u) *WrapModeRepeat {
	return &WrapModeRepeat{
		size:  size,
		add:   size * (0x3FFFFFFF / size),
		value: 0,
	}
}

// Call wraps the given coordinate using repeat mode.
func (w *WrapModeRepeat) Call(v int) basics.Int32u {
	w.value = (basics.Int32u(v) + w.add) % w.size
	return w.value
}

// Inc increments the current position with wrapping.
func (w *WrapModeRepeat) Inc() basics.Int32u {
	w.value++
	if w.value >= w.size {
		w.value = 0
	}
	return w.value
}

// WrapModeRepeatPow2 implements optimized repeat wrapping for power-of-2 sizes.
// Uses bitmasking instead of modulo for better performance.
type WrapModeRepeatPow2 struct {
	mask  basics.Int32u
	value basics.Int32u
}

// NewWrapModeRepeatPow2 creates a new power-of-2 repeat wrap mode.
func NewWrapModeRepeatPow2(size basics.Int32u) *WrapModeRepeatPow2 {
	mask := basics.Int32u(1)
	for mask < size {
		mask = (mask << 1) | 1
	}
	mask >>= 1

	return &WrapModeRepeatPow2{
		mask:  mask,
		value: 0,
	}
}

// Call wraps the given coordinate using power-of-2 repeat mode.
func (w *WrapModeRepeatPow2) Call(v int) basics.Int32u {
	w.value = basics.Int32u(v) & w.mask
	return w.value
}

// Inc increments the current position with wrapping.
func (w *WrapModeRepeatPow2) Inc() basics.Int32u {
	w.value++
	if w.value > w.mask {
		w.value = 0
	}
	return w.value
}

// WrapModeRepeatAutoPow2 automatically chooses between regular and power-of-2 repeat modes.
type WrapModeRepeatAutoPow2 struct {
	size  basics.Int32u
	add   basics.Int32u
	mask  basics.Int32u
	value basics.Int32u
}

// NewWrapModeRepeatAutoPow2 creates a new auto-detecting repeat wrap mode.
func NewWrapModeRepeatAutoPow2(size basics.Int32u) *WrapModeRepeatAutoPow2 {
	add := size * (0x3FFFFFFF / size)
	mask := basics.Int32u(0)

	// Check if size is power of 2
	if (size & (size - 1)) == 0 {
		mask = size - 1
	}

	return &WrapModeRepeatAutoPow2{
		size:  size,
		add:   add,
		mask:  mask,
		value: 0,
	}
}

// Call wraps the given coordinate using auto-detected repeat mode.
func (w *WrapModeRepeatAutoPow2) Call(v int) basics.Int32u {
	if w.mask != 0 {
		w.value = basics.Int32u(v) & w.mask
	} else {
		w.value = (basics.Int32u(v) + w.add) % w.size
	}
	return w.value
}

// Inc increments the current position with wrapping.
func (w *WrapModeRepeatAutoPow2) Inc() basics.Int32u {
	w.value++
	if w.value >= w.size {
		w.value = 0
	}
	return w.value
}

// WrapModeReflect implements texture wrapping with reflect (mirror) mode.
type WrapModeReflect struct {
	size  basics.Int32u
	size2 basics.Int32u
	add   basics.Int32u
	value basics.Int32u
}

// NewWrapModeReflect creates a new reflect wrap mode for the given size.
func NewWrapModeReflect(size basics.Int32u) *WrapModeReflect {
	size2 := size * 2
	return &WrapModeReflect{
		size:  size,
		size2: size2,
		add:   size2 * (0x3FFFFFFF / size2),
		value: 0,
	}
}

// Call wraps the given coordinate using reflect mode.
func (w *WrapModeReflect) Call(v int) basics.Int32u {
	w.value = (basics.Int32u(v) + w.add) % w.size2
	if w.value >= w.size {
		return w.size2 - w.value - 1
	}
	return w.value
}

// Inc increments the current position with wrapping.
func (w *WrapModeReflect) Inc() basics.Int32u {
	w.value++
	if w.value >= w.size2 {
		w.value = 0
	}
	if w.value >= w.size {
		return w.size2 - w.value - 1
	}
	return w.value
}

// WrapModeReflectPow2 implements optimized reflect wrapping for power-of-2 sizes.
type WrapModeReflectPow2 struct {
	size  basics.Int32u
	mask  basics.Int32u
	value basics.Int32u
}

// NewWrapModeReflectPow2 creates a new power-of-2 reflect wrap mode.
func NewWrapModeReflectPow2(size basics.Int32u) *WrapModeReflectPow2 {
	mask := basics.Int32u(1)
	actualSize := basics.Int32u(1)

	for mask < size {
		mask = (mask << 1) | 1
		actualSize <<= 1
	}

	return &WrapModeReflectPow2{
		size:  actualSize,
		mask:  mask,
		value: 0,
	}
}

// Call wraps the given coordinate using power-of-2 reflect mode.
func (w *WrapModeReflectPow2) Call(v int) basics.Int32u {
	w.value = basics.Int32u(v) & w.mask
	if w.value >= w.size {
		return w.mask - w.value
	}
	return w.value
}

// Inc increments the current position with wrapping.
func (w *WrapModeReflectPow2) Inc() basics.Int32u {
	w.value++
	w.value &= w.mask
	if w.value >= w.size {
		return w.mask - w.value
	}
	return w.value
}

// WrapModeReflectAutoPow2 automatically chooses between regular and power-of-2 reflect modes.
type WrapModeReflectAutoPow2 struct {
	size  basics.Int32u
	size2 basics.Int32u
	add   basics.Int32u
	mask  basics.Int32u
	value basics.Int32u
}

// NewWrapModeReflectAutoPow2 creates a new auto-detecting reflect wrap mode.
func NewWrapModeReflectAutoPow2(size basics.Int32u) *WrapModeReflectAutoPow2 {
	size2 := size * 2
	add := size2 * (0x3FFFFFFF / size2)
	mask := basics.Int32u(0)

	// Check if size2 is power of 2
	if (size2 & (size2 - 1)) == 0 {
		mask = size2 - 1
	}

	return &WrapModeReflectAutoPow2{
		size:  size,
		size2: size2,
		add:   add,
		mask:  mask,
		value: 0,
	}
}

// Call wraps the given coordinate using auto-detected reflect mode.
func (w *WrapModeReflectAutoPow2) Call(v int) basics.Int32u {
	if w.mask != 0 {
		w.value = basics.Int32u(v) & w.mask
	} else {
		w.value = (basics.Int32u(v) + w.add) % w.size2
	}

	if w.value >= w.size {
		return w.size2 - w.value - 1
	}
	return w.value
}

// Inc increments the current position with wrapping.
func (w *WrapModeReflectAutoPow2) Inc() basics.Int32u {
	w.value++
	if w.value >= w.size2 {
		w.value = 0
	}
	if w.value >= w.size {
		return w.size2 - w.value - 1
	}
	return w.value
}
