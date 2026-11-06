package controller

import (
	"fmt"
	"testing"
)

var biodataSpec = crudSpec{
	base: "/m_biodata",
	createBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"fullname":"Name","mobilePhone":"0811","isDelete":false}`, id)
	},
	updateBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"fullname":"Changed","mobilePhone":"0822","isDelete":false}`, id)
	},
}

func TestMBiodata_CRUDSuite(t *testing.T) {
	runCRUDSuite(t, biodataSpec)
}
