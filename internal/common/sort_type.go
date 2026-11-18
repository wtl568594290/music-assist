package common

import "fmt"

type SortType string

const (
	SortByName        SortType = "Name"
	SortByNameDesc    SortType = "NameDesc"
	SortByModTime     SortType = "ModTime"
	SortByModTimeDesc SortType = "ModTimeDesc"
	SortByCreated     SortType = "Created"
	SortByCreatedDesc SortType = "CreatedDesc"
)

var validSortTypes = map[string]SortType{
	"Name":        SortByName,
	"NameDesc":    SortByNameDesc,
	"ModTime":     SortByModTime,
	"ModTimeDesc": SortByModTimeDesc,
	"Created":     SortByCreated,
	"CreatedDesc": SortByCreatedDesc,
}

func GetSortTypeList() []SortType {
	keys := make([]SortType, 0, len(validSortTypes))
	for _, st := range validSortTypes {
		keys = append(keys, st)
	}
	return keys
}

func ParseSortType(s string) (SortType, error) {
	st, ok := validSortTypes[s]
	if !ok {
		return "", fmt.Errorf("invalid SortType: %s", s)
	}
	return st, nil
}
