package generator

import "github.com/aykhans/bsky-feedgen/pkg/utils"

type Users map[string]bool

func (u Users) IsValid(did string) *bool {
	isValid, ok := u[did]
	if ok == false {
		return nil
	}

	return utils.ToPtr(isValid)
}
