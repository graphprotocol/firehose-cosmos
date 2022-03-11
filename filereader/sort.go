package filereader

import (
	"io/fs"
)

type sortFilesByName []fs.DirEntry

func (f sortFilesByName) Len() int {
	return len(f)
}

func (f sortFilesByName) Less(i, j int) bool {
	return f[i].Name() < f[j].Name()
}

func (f sortFilesByName) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
