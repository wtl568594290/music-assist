package common

type FileChangeFlag struct {
	ModTime int64
	Size    int64
	Hash    string
}

type FileInfo struct {
	Path     string
	Dir      string
	BaseName string
	Ext      string
	// ModTime  int64
	Created int64
	// Size     int64
	// Hash string
	FileChangeFlag
}
