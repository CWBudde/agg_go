package basics

import (
	"reflect"
	"testing"
	"unsafe"
)

// Test types for allocator testing
type TestStruct struct {
	X, Y int
	Name string
}

func TestPodAllocator_Allocate(t *testing.T) {
	tests := []struct {
		name string
		num  int
		want int // expected length
	}{
		{"Allocate zero", 0, 0},
		{"Allocate negative", -1, 0},
		{"Allocate one", 1, 1},
		{"Allocate multiple", 10, 10},
		{"Allocate large", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocator := NewPodAllocator[int]()
			got := allocator.Allocate(tt.num)

			if len(got) != tt.want {
				t.Errorf("PodAllocator.Allocate() length = %v, want %v", len(got), tt.want)
			}

			if tt.num > 0 {
				if cap(got) < tt.want {
					t.Errorf("PodAllocator.Allocate() capacity = %v, want at least %v", cap(got), tt.want)
				}

				// Test that slice is properly zero-initialized
				for i, v := range got {
					if v != 0 {
						t.Errorf("PodAllocator.Allocate() slice[%d] = %v, want 0", i, v)
					}
				}
			}
		})
	}
}

func TestPodAllocator_DifferentTypes(t *testing.T) {
	t.Run("int allocator", func(t *testing.T) {
		allocator := NewPodAllocator[int]()
		slice := allocator.Allocate(5)
		if len(slice) != 5 {
			t.Errorf("Expected length 5, got %d", len(slice))
		}
	})

	t.Run("float64 allocator", func(t *testing.T) {
		allocator := NewPodAllocator[float64]()
		slice := allocator.Allocate(3)
		if len(slice) != 3 {
			t.Errorf("Expected length 3, got %d", len(slice))
		}
		for i, v := range slice {
			if v != 0.0 {
				t.Errorf("slice[%d] = %v, want 0.0", i, v)
			}
		}
	})

	t.Run("struct allocator", func(t *testing.T) {
		allocator := NewPodAllocator[TestStruct]()
		slice := allocator.Allocate(2)
		if len(slice) != 2 {
			t.Errorf("Expected length 2, got %d", len(slice))
		}
		for i, v := range slice {
			if v.X != 0 || v.Y != 0 || v.Name != "" {
				t.Errorf("slice[%d] = %+v, want zero value", i, v)
			}
		}
	})

	t.Run("byte allocator", func(t *testing.T) {
		allocator := NewPodAllocator[byte]()
		slice := allocator.Allocate(100)
		if len(slice) != 100 {
			t.Errorf("Expected length 100, got %d", len(slice))
		}
	})
}

func TestPodAllocator_Deallocate(t *testing.T) {
	allocator := NewPodAllocator[int]()
	slice := allocator.Allocate(10)

	// Deallocate should not panic and should be a no-op
	allocator.Deallocate(slice, 10)

	// Slice should still be accessible (GC handles the cleanup)
	if len(slice) != 10 {
		t.Errorf("After deallocate, slice length = %v, want 10", len(slice))
	}
}

func TestObjAllocator_Allocate(t *testing.T) {
	t.Run("int allocator", func(t *testing.T) {
		allocator := NewObjAllocator[int]()
		ptr := allocator.Allocate()

		if ptr == nil {
			t.Error("ObjAllocator.Allocate() returned nil")
		}

		// Test that the value is zero-initialized
		if *ptr != 0 {
			t.Errorf("ObjAllocator.Allocate() *ptr = %v, want 0", *ptr)
		}

		// Test that we can modify the value
		*ptr = 42
		if *ptr != 42 {
			t.Errorf("After assignment, *ptr = %v, want 42", *ptr)
		}
	})

	t.Run("struct allocator", func(t *testing.T) {
		allocator := NewObjAllocator[TestStruct]()
		ptr := allocator.Allocate()

		if ptr == nil {
			t.Error("ObjAllocator.Allocate() returned nil")
		}

		// Test zero initialization
		if ptr.X != 0 || ptr.Y != 0 || ptr.Name != "" {
			t.Errorf("ObjAllocator.Allocate() ptr = %+v, want zero value", ptr)
		}

		// Test modification
		ptr.X = 10
		ptr.Y = 20
		ptr.Name = "test"

		if ptr.X != 10 || ptr.Y != 20 || ptr.Name != "test" {
			t.Errorf("After modification, ptr = %+v, want {X:10 Y:20 Name:test}", ptr)
		}
	})
}

func TestObjAllocator_Deallocate(t *testing.T) {
	allocator := NewObjAllocator[int]()
	ptr := allocator.Allocate()
	*ptr = 42

	// Deallocate should not panic and should be a no-op
	allocator.Deallocate(ptr)

	// Pointer should still be accessible (GC handles the cleanup)
	if *ptr != 42 {
		t.Errorf("After deallocate, *ptr = %v, want 42", *ptr)
	}
}

func TestGlobalAllocators(t *testing.T) {
	t.Run("DefaultPodAllocator", func(t *testing.T) {
		slice := DefaultPodAllocator.Allocate(5)
		if len(slice) != 5 {
			t.Errorf("DefaultPodAllocator length = %v, want 5", len(slice))
		}
	})

	t.Run("DefaultObjAllocator", func(t *testing.T) {
		ptr := DefaultObjAllocator.Allocate()
		if ptr == nil {
			t.Error("DefaultObjAllocator returned nil")
		}
	})
}

func TestAllocatorInterfaces(t *testing.T) {
	t.Run("PodAllocator implements AllocatorInterface", func(t *testing.T) {
		var allocator AllocatorInterface[int] = NewPodAllocator[int]()
		slice := allocator.Allocate(5)
		if len(slice) != 5 {
			t.Errorf("Interface allocation length = %v, want 5", len(slice))
		}
		allocator.Deallocate(slice, 5)
	})

	t.Run("ObjAllocator implements ObjectAllocatorInterface", func(t *testing.T) {
		var allocator ObjectAllocatorInterface[int] = NewObjAllocator[int]()
		ptr := allocator.Allocate()
		if ptr == nil {
			t.Error("Interface object allocation returned nil")
		}
		allocator.Deallocate(ptr)
	})
}

func TestAllocatorTypeSafety(t *testing.T) {
	// Test that allocators are type-safe
	intAllocator := NewPodAllocator[int]()
	intSlice := intAllocator.Allocate(5)

	stringAllocator := NewPodAllocator[string]()
	stringSlice := stringAllocator.Allocate(3)

	// Verify types
	if reflect.TypeOf(intSlice).Elem() != reflect.TypeOf(int(0)) {
		t.Error("Int allocator should produce int slice")
	}

	if reflect.TypeOf(stringSlice).Elem() != reflect.TypeOf("") {
		t.Error("String allocator should produce string slice")
	}
}

func TestMemoryAlignment(t *testing.T) {
	// Test that allocated memory is properly aligned for the type
	t.Run("int alignment", func(t *testing.T) {
		allocator := NewPodAllocator[int]()
		slice := allocator.Allocate(1)
		if len(slice) > 0 {
			addr := uintptr(unsafe.Pointer(&slice[0]))
			if addr%unsafe.Alignof(int(0)) != 0 {
				t.Error("Int slice is not properly aligned")
			}
		}
	})

	t.Run("struct alignment", func(t *testing.T) {
		allocator := NewPodAllocator[TestStruct]()
		slice := allocator.Allocate(1)
		if len(slice) > 0 {
			addr := uintptr(unsafe.Pointer(&slice[0]))
			if addr%unsafe.Alignof(TestStruct{}) != 0 {
				t.Error("Struct slice is not properly aligned")
			}
		}
	})
}

func TestNilSafety(t *testing.T) {
	t.Run("deallocate nil slice", func(t *testing.T) {
		allocator := NewPodAllocator[int]()
		// Should not panic
		allocator.Deallocate(nil, 0)
	})

	t.Run("deallocate nil pointer", func(t *testing.T) {
		allocator := NewObjAllocator[int]()
		// Should not panic
		allocator.Deallocate(nil)
	})
}

// Benchmark tests to compare allocator performance with direct Go allocation

func BenchmarkPodAllocator_Small(b *testing.B) {
	allocator := NewPodAllocator[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slice := allocator.Allocate(10)
		allocator.Deallocate(slice, 10)
	}
}

func BenchmarkDirectAllocation_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice := make([]int, 10)
		_ = slice // Use slice to prevent optimization
	}
}

func BenchmarkPodAllocator_Medium(b *testing.B) {
	allocator := NewPodAllocator[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slice := allocator.Allocate(1000)
		allocator.Deallocate(slice, 1000)
	}
}

func BenchmarkDirectAllocation_Medium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice := make([]int, 1000)
		_ = slice
	}
}

func BenchmarkPodAllocator_Large(b *testing.B) {
	allocator := NewPodAllocator[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slice := allocator.Allocate(100000)
		allocator.Deallocate(slice, 100000)
	}
}

func BenchmarkDirectAllocation_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice := make([]int, 100000)
		_ = slice
	}
}

func BenchmarkObjAllocator(b *testing.B) {
	allocator := NewObjAllocator[TestStruct]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ptr := allocator.Allocate()
		allocator.Deallocate(ptr)
	}
}

func BenchmarkDirectObjectAllocation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ptr := new(TestStruct)
		_ = ptr
	}
}

func BenchmarkPodAllocator_DifferentTypes(b *testing.B) {
	b.Run("int", func(b *testing.B) {
		allocator := NewPodAllocator[int]()
		for i := 0; i < b.N; i++ {
			slice := allocator.Allocate(100)
			allocator.Deallocate(slice, 100)
		}
	})

	b.Run("float64", func(b *testing.B) {
		allocator := NewPodAllocator[float64]()
		for i := 0; i < b.N; i++ {
			slice := allocator.Allocate(100)
			allocator.Deallocate(slice, 100)
		}
	})

	b.Run("byte", func(b *testing.B) {
		allocator := NewPodAllocator[byte]()
		for i := 0; i < b.N; i++ {
			slice := allocator.Allocate(100)
			allocator.Deallocate(slice, 100)
		}
	})

	b.Run("struct", func(b *testing.B) {
		allocator := NewPodAllocator[TestStruct]()
		for i := 0; i < b.N; i++ {
			slice := allocator.Allocate(100)
			allocator.Deallocate(slice, 100)
		}
	})
}

func BenchmarkAllocatorReuse(b *testing.B) {
	allocator := NewPodAllocator[int]()

	b.Run("new allocator each time", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			newAllocator := NewPodAllocator[int]()
			slice := newAllocator.Allocate(100)
			newAllocator.Deallocate(slice, 100)
		}
	})

	b.Run("reuse allocator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := allocator.Allocate(100)
			allocator.Deallocate(slice, 100)
		}
	})
}
