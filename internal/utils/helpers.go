// utils/helpers.go
package utils

func AppendIfNotExists(slice []string, item string) []string {
	for _, existing := range slice {
		if existing == item {
			return slice
		}
	}
	return append(slice, item)
}
