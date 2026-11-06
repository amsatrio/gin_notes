package controller

import (
	"fmt"
	"testing"
)

var userSpec = crudSpec{
	base: "/m_user",
	createBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"biodataId":1,"roleId":1,"email":"u%[1]d@example.com","password":"pw",`+nestedBiodataJSON+`,`+nestedRoleJSON+`}`, id)
	},
	updateBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"biodataId":1,"roleId":1,"email":"changed@example.com","password":"pw2",`+nestedBiodataJSON+`,`+nestedRoleJSON+`}`, id)
	},
}

func TestMUser_CRUDSuite(t *testing.T) {
	runCRUDSuite(t, userSpec)
}
