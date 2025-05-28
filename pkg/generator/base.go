package generator

import "github.com/aykhans/bsky-feedgen/pkg/utils"

type Users map[string]bool

// IsValid checks if a given DID exists in the Users map and returns its validity status.
//
// Parameters:
//
//	did: The Decentralized Identifier string to check
//
// Returns:
//   - *bool: A pointer to the validity status if the DID exists in the map
//   - nil: If the DID does not exist in the map
func (u Users) IsValid(did string) *bool {
	isValid, ok := u[did]
	if ok == false {
		return nil
	}

	return utils.ToPtr(isValid)
}

// GetValidUsers returns a slice of DIDs that are marked as valid in the Users map.
//
// Returns:
//   - []string: A slice of valid DIDs, limited by the specified parameters
func (u Users) GetValidUsers() []string {
	validUsers := make([]string, 0)

	for did, isValid := range u {
		if isValid {
			validUsers = append(validUsers, did)
		}
	}

	return validUsers
}

// GetInvalidUsers returns a slice of DIDs that are marked as invalid in the Users map.
//
// Returns:
//   - []string: A slice of invalid DIDs, limited by the specified parameters
func (u Users) GetInvalidUsers() []string {
	invalidUsers := make([]string, 0)

	for did, isValid := range u {
		if !isValid {
			invalidUsers = append(invalidUsers, did)
		}
	}

	return invalidUsers
}

// GetAll returns a slice of all DIDs in the Users map, regardless of validity status.
//
// Returns:
//   - []string: A slice containing all DIDs in the map
func (u Users) GetAll() []string {
	allUsers := make([]string, 0, len(u))

	for did := range u {
		allUsers = append(allUsers, did)
	}

	return allUsers
}

type Langs map[string]bool

// IsExistsAny checks if any of the given language codes exist in the Langs map.
//
// Parameters:
//   - langs: A slice of language code strings to check for existence
//
// Returns:
//   - bool: true if at least one language code from the input slice exists in the map,
//           false if none of the provided language codes exist
func (l Langs) IsExistsAny(langs []string) bool {
	for _, lang := range langs {
		if _, ok := l[lang]; ok {
			return true
		}
	}

	return false
}
