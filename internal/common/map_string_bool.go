package common

type MapStringBool map[string]bool

func NewMapStringBool() MapStringBool {
	return make(MapStringBool)
}

func (m MapStringBool) ToSlice() []Source {
	slice := make([]Source, 0, len(m))
	for path, isDir := range m {
		slice = append(slice, Source{
			Path:  path,
			IsDir: isDir,
		})
	}
	return slice
}

func (m MapStringBool) Add(path string, isDir bool) bool {
	_, ok := m[path]
	if ok {
		return false
	}
	m[path] = isDir
	return true
}
