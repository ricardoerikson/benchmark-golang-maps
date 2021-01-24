package main

// MapWithInterface uses interfaces
func MapWithInterface(Entries, Capacity int) {
	m := make(map[int]interface{}, Capacity)
	for i := 1; i <= Entries; i++ {
		m[i] = nil
	}
}

var m map[int]struct{}

// MapWithEmptyStruct uses structs
func MapWithEmptyStruct(Entries, Capacity int) {
	m := make(map[int]struct{}, Capacity)
	for i := 1; i <= Entries; i++ {
		m[i] = struct{}{}
	}
}
