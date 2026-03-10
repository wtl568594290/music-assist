package app

type Source struct {
	Path  string
	IsDir bool
}

type MapStringStruct map[string]struct{}
type MapStringBool map[string]bool

func (m MapStringStruct) ToSlice() []string {
	slice := make([]string, 0, len(m))
	for path := range m {
		slice = append(slice, path)
	}
	return slice
}

func (m MapStringStruct) Add(path string) bool {
	_, ok := m[path]
	if ok {
		return false
	}
	m[path] = struct{}{}
	return true
}

func NewMapStringStruct() MapStringStruct {
	return make(MapStringStruct)
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

func NewMapStringBool() MapStringBool {
	return make(MapStringBool)
}
