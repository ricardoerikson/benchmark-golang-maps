package main

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

var benchmarks = []struct {
	Entries         int
	InitialCapacity int
}{
	{7, 0},
	{8, 0},
	{9, 0},
	{16, 16},
	{16, 32},
	{16, 100},
	{100, 0},
	{100, 128},
}

func Benchmark_Interface(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MapWithInterface(16, 100)
	}
}

func Benchmark_EmptyStruct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MapWithEmptyStruct(100, 0)
	}
}

// Test_EmptyStructValueSize shows that empty struct value has size zero.
func Test_EmptyStructValueSize(t *testing.T) {
	var myStruct struct{} = struct{}{}
	size := unsafe.Sizeof(myStruct)
	assert.True(t, uintptr(0) == size)
}

// Test_NilInterfaceValueSize shows that an interface with nil value does not have size zero.
func Test_NilInterfaceValueSize(t *testing.T) {
	var myInterface interface{} = nil
	size := unsafe.Sizeof(myInterface)
	assert.True(t, uintptr(0) != size)
}

// Benchmark_EmptyStructTDB is table driven benchmark for map with empty struct
func Benchmark_EmptyStructTDB(b *testing.B) {
	for _, bm := range benchmarks {
		desc := fmt.Sprintf("Entries: %3d, InitialCapacity: %3d\n", bm.Entries, bm.InitialCapacity)
		b.Run(desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				MapWithEmptyStruct(bm.Entries, bm.InitialCapacity)
			}
		})
	}
}

// Benchmark_InterfaceTDB is table driven benchmark for map with interface
func Benchmark_InterfaceTDB(b *testing.B) {
	for _, bm := range benchmarks {
		desc := fmt.Sprintf("Entries: %3d, InitialCapacity: %3d\n", bm.Entries, bm.InitialCapacity)
		b.Run(desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				MapWithInterface(bm.Entries, bm.InitialCapacity)
			}
		})
	}
}
