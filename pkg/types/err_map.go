package types

type ErrMap map[string]error

func (ErrMap ErrMap) ToStringMap() map[string]string {
	if len(ErrMap) == 0 {
		return nil
	}

	stringMap := make(map[string]string)
	for key, err := range ErrMap {
		if err != nil {
			stringMap[key] = err.Error()
		}
	}

	return stringMap
}
