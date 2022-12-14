package testdomain

import (
	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
)

// DomainLevel0base only contains ObjLevel0 which is a collection of fields that makes good starting points for feature support.
func DomainLevel0base() *seed.Domain {
	return must.V(seed.NewDomain(
		seed.Thing{
			Name: "test_level_0_base",
		},
		ObjLevel0(),
	))
}

// DomainLevel0All is a collection of objects that test for basic feature support.
func DomainLevel0All() *seed.Domain {
	return must.V(seed.NewDomain(
		seed.Thing{
			Name: "test_level_0_all",
		},
		append(DomainLevel0base().Objects.Values(), ObjLevel0List())...,
	))
}

// DomainLevel1 is a collection of objects that test for Level 1 feature support.
func DomainLevel1() *seed.Domain {
	return must.V(seed.NewDomain(
		seed.Thing{
			Name: "test_level_1",
		},
		append(DomainLevel0All().Objects.Values(), ObjTimeStampCommon())...,
	))
}
