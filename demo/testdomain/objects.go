package testdomain

import (
	"github.com/xiegeo/must"

	"github.com/xiegeo/seed"
)

// ObjLevel0 is a collection of fields that makes good starting points for feature support.
func ObjLevel0() *seed.Object {
	return &seed.Object{
		Thing: seed.Thing{
			Name: "level_0",
		},
		FieldGroup: seed.FieldGroup{
			Fields: must.V(seed.NewFields(
				TextLineField(),
				Bytes(),
				Bool(),
				DateTimeSec(),
				JSInteger(),
			)),
			Identities: []seed.Identity{{
				Fields: []seed.CodeName{TextLineField().Name},
			}},
		},
	}
}

// ObjLevel0List is all list extension of ObjLevel0
func ObjLevel0List() *seed.Object {
	setting := seed.ListSetting{
		MinLength: 0,
		MaxLength: 5,
		IsOrdered: true,
		IsUnique:  false,
	}
	return &seed.Object{
		Thing: seed.Thing{
			Name: "level_0_list",
		},
		FieldGroup: seed.FieldGroup{
			Fields: must.V(ObjLevel0().Fields.NewMap(func(f *seed.Field) (*seed.Field, error) {
				return ListOf(f, setting), nil
			})),
		},
	}
}

// ObjTimeStampCommon is a collection of fields that covers common timestamps options
func ObjTimeStampCommon() *seed.Object {
	return &seed.Object{
		Thing: seed.Thing{
			Name: "time_stamp_common",
		},
		FieldGroup: seed.FieldGroup{
			Fields: must.V(seed.NewFields(
				Date(),
				DateTimeSec(),
				DateTimeMill(),
				WithTimeZone(DateTimeMicro()),
				DateTimeNano(),
			)),
			Identities: []seed.Identity{{
				Fields: []seed.CodeName{DateTimeSec().Name},
			}},
		},
	}
}
