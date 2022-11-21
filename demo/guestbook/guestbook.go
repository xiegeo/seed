package main

import (
	"math/big"
	"time"

	"golang.org/x/text/language"

	"github.com/xiegeo/seed"
)

func Domain() seed.Domain {
	return seed.Domain{
		Thing: seed.Thing{
			Name: "guestbook",
		},
		Objects: []seed.Object{
			{
				Thing: seed.Thing{
					Name: "guest",
				},
				FieldProperties: seed.FieldProperties{
					Fields: []seed.Field{
						TimeField(),
						NameField(),
						// EventsField(), // todo: list and ref types
						NumberOfGuestsField(),
						// ContactField(), // todo: combination type
						NoteField(),
					},
					Identities: []seed.Identity{{
						Fields: []seed.CodeName{TimeField().Name, NameField().Name},
					}},
				},
			}, {
				Thing: seed.Thing{
					Name: "event",
				},
				FieldProperties: seed.FieldProperties{
					Fields: []seed.Field{
						StartTimeField(),
						EndTimeField(),
						// EventNameField(), // todo:i18n
						PublishField(),
						MaxNumberOfGuestsField(),
						// EventDescriptionField(), // todo:i18n
					},
					Identities: []seed.Identity{{
						Fields: []seed.CodeName{EventNameField().Name},
					}},
					Ranges: []seed.Range{{
						Start:           StartTimeField().Name,
						End:             EndTimeField().Name,
						IncludeEndValue: false,
					}},
				},
			},
		},
	}
}

func NameField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "name",
			Label: seed.I18n[string]{
				language.English: "Name",
				language.Chinese: "名称",
			},
		},
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MinCodePoints: 2,
			MaxCodePoints: 100,
			IsSingleLine:  true,
		},
	}
}

func EventNameField() seed.Field {
	f := NameField()
	f.Label = seed.I18n[string]{
		language.English: "Event Name",
		language.Chinese: "活动名称",
	}
	f.IsI18n = true
	return f
}

func NumberOfGuestsField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "number_of_guests",
			Label: seed.I18n[string]{
				language.English: "Number of Guests",
				language.Chinese: "人数",
			},
		},
		FieldType: seed.Integer,
		FieldTypeSetting: seed.IntegerSetting{
			Min: big.NewInt(1),
			Max: big.NewInt(99),
		},
	}
}

func MaxNumberOfGuestsField() seed.Field {
	f := NumberOfGuestsField()
	f.Name = "max_number_of_guests"
	f.Label = seed.I18n[string]{
		language.English: "Compacity",
		language.Chinese: "人数上限",
	}
	f.FieldTypeSetting = seed.IntegerSetting{
		Min: big.NewInt(1),
		Max: big.NewInt(99999),
	}
	return f
}

func TimeField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "time",
			Label: seed.I18n[string]{
				language.English: "Time",
				language.Chinese: "时间",
			},
		},
		FieldType: seed.TimeStamp,
		FieldTypeSetting: seed.TimeStampSetting{
			Min:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			Max:   time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC),
			Scale: time.Second,
		},
	}
}

func StartTimeField() seed.Field {
	f := TimeField()
	f.Name = "start_time"
	f.Label = seed.I18n[string]{
		language.English: "Start Time",
		language.Chinese: "开始时间",
	}
	f.FieldTypeSetting = seed.TimeStampSetting{
		Min:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Max:   time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC),
		Scale: time.Minute * 5,
	}
	return f
}

func EndTimeField() seed.Field {
	f := StartTimeField()
	f.Name = "end_time"
	f.Label = seed.I18n[string]{
		language.English: "End Time",
		language.Chinese: "结束时间",
	}
	return f
}

func ContactField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "contact",
			Label: seed.I18n[string]{
				language.English: "contract",
				language.Chinese: "联系方式",
			},
		},
		FieldType: seed.Combination,
		FieldTypeSetting: seed.CombinationSetting{
			Fields: []seed.Field{
				PhoneNumberField(),
				EmailField(),
			},
		},
	}
}

func PhoneNumberField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "phone_number",
			Label: seed.I18n[string]{
				language.English: "Phone Number",
				language.Chinese: "电话",
			},
		},
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MaxCodePoints: 20,
			IsSingleLine:  true,
		},
		Nullable: true,
	}
}

func EmailField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "email",
			Label: seed.I18n[string]{
				language.English: "eMail",
				language.Chinese: "邮箱",
			},
		},
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MaxCodePoints: 50,
			IsSingleLine:  true,
		},
		Nullable: true,
	}
}

func NoteField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "note",
			Label: seed.I18n[string]{
				language.English: "Note",
				language.Chinese: "标注",
			},
		},
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MaxCodePoints: 500,
			IsSingleLine:  false,
		},
	}
}

func EventDescriptionField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "description",
			Label: seed.I18n[string]{
				language.English: "Event Details",
				language.Chinese: "活动描述",
			},
		},
		FieldType: seed.String,
		FieldTypeSetting: seed.StringSetting{
			MaxCodePoints: 5000,
			IsSingleLine:  false,
		},
		IsI18n: true,
	}
}

func EventsField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "events",
			Label: seed.I18n[string]{
				language.English: "Events",
				language.Chinese: "活动",
			},
		},
		FieldType: seed.List,
		FieldTypeSetting: seed.ListSetting{
			MaxLength: 20,
			IsOrdered: true,
			IsUnique:  true,
			ItemType:  seed.Reference,
			ItemTypeSetting: seed.ReferenceSetting{
				Target: "event",
			},
		},
	}
}

func PublishField() seed.Field {
	return seed.Field{
		Thing: seed.Thing{
			Name: "publish",
			Label: seed.I18n[string]{
				language.English: "Publish",
				language.Chinese: "发布",
			},
		},
		FieldType:        seed.Boolean,
		FieldTypeSetting: seed.BooleanSetting{},
	}
}
