package bar

import "parti/api"

// Type ...
type Type struct {
	ID         int64              `json:"id" comment:"уникальный номер" validate:"required"`
	Name       string             `json:"name" comment:"имя" validate:"required"`
	Groups     []GroupType        `json:"groups" comment:"группы" validate:"required,dive,required"`
	NullString api.EmpNullString  `json:"nullString" validate:"required"`
	NullInt    api.EmpNullInt64   `json:"nullInt"`
	NullFloat  api.EmpNullFloat64 `json:"nullFloat" validate:"required"`
	NullBool   api.EmpNullBool    `json:"nullBool" validate:"required"`
}

// GroupType ...
type GroupType struct {
	ID      int64  `json:"id" comment:"уникальный номер" validate:"required"`
	GrpName string `json:"grp_name" comment:"имя" validate:"required"`
	GrpAge  int    `json:"grp_age"`
}

type OutType struct {
	ID         int64              `json:"id" comment:"уникальный номер"`
	Name       string             `json:"name" comment:"имя"`
	Port       string             `json:"port" comment:"порт"`
	Groups     []GroupType        `json:"groups" comment:"группы"`
	NullString api.EmpNullString  `json:"nullString"`
	NullInt    api.EmpNullInt64   `json:"nullInt"`
	NullFloat  api.EmpNullFloat64 `json:"nullFloat"`
	NullBool   api.EmpNullBool    `json:"nullBool"`
}
