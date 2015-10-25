package pkgmap // import "go.pedge.io/pkg/map"

import "strconv"

// StringStringMap is a map from strings to strings.
type StringStringMap map[string]string

// GetString gets the string if it is present, or "" otherwise.
func (s StringStringMap) GetString(key string) (string, error) {
	value, ok := s[key]
	if !ok {
		return "", nil
	}
	return value, nil
}

// GetInt32 gets the int32 if it is present, or 0 otherwise.
func (s StringStringMap) GetInt32(key string) (int32, error) {
	value, ok := s[key]
	if !ok {
		return 0, nil
	}
	parsedValue, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsedValue), nil
}

// GetInt64 gets the int64 if it is present, or 0 otherwise.
func (s StringStringMap) GetInt64(key string) (int64, error) {
	value, ok := s[key]
	if !ok {
		return 0, nil
	}
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return int64(parsedValue), nil
}

// GetUint32 gets the uint32 if it is present, or 0 otherwise.
func (s StringStringMap) GetUint32(key string) (uint32, error) {
	value, ok := s[key]
	if !ok {
		return 0, nil
	}
	parsedValue, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsedValue), nil
}

// GetUint64 gets the uint64 if it is present, or 0 otherwise.
func (s StringStringMap) GetUint64(key string) (uint64, error) {
	value, ok := s[key]
	if !ok {
		return 0, nil
	}
	parsedValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(parsedValue), nil
}

// GetBool gets the bool if it is present, or 0 otherwise.
func (s StringStringMap) GetBool(key string) (bool, error) {
	value, ok := s[key]
	if !ok {
		return false, nil
	}
	return strconv.ParseBool(value)
}
