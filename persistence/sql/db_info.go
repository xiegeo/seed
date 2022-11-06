package sql

import "github.com/xiegeo/seed"

type domainInfo struct {
	seed.Domain
	objectMap map[seed.CodeName]objectInfo
}

func domainInfoFromDomain(d seed.Domain) domainInfo {
	objectMap := make(map[seed.CodeName]objectInfo, len(d.Objects))
	for _, ob := range d.Objects {
		objectMap[ob.Name] = objectInfoFromObject(ob)
	}
	return domainInfo{
		Domain:    d,
		objectMap: objectMap,
	}
}

type objectInfo struct {
	seed.Object
	fieldMap map[seed.CodeName]fieldInfo
}

func objectInfoFromObject(ob seed.Object) objectInfo {
	fieldMap := make(map[seed.CodeName]fieldInfo, len(ob.Fields))
	for _, f := range ob.Fields {
		fieldMap[f.Name] = fieldInfo{
			Field: f,
		}
	}
	return objectInfo{
		Object:   ob,
		fieldMap: fieldMap,
	}
}

type fieldInfo struct {
	seed.Field
}
