package util

import "os"

type MemoryOffsetList struct {
	buffer Buffer
}

func NewMemoryOffsetList() *MemoryOffsetList {
	return new(MemoryOffsetList)
}

func (mol *MemoryOffsetList) AddOffset(name string, offset string) {
	mol.buffer.WriteString(name + " " + offset + "\n")
}

func (mol *MemoryOffsetList) WriteMemoryOffsetFile() {
	file := "memory_offsets.txt"
	newFile, _ := os.Create(file)
	defer newFile.Close()

	newFile.Write(mol.buffer.Bytes())
}
