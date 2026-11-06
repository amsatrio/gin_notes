package controller

import (
	"fmt"
	"testing"
)

var roleSpec = crudSpec{
	base: "/m_role",
	createBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"name":"Role","code":"R%[1]d","level":2,"isDelete":false}`, id)
	},
	updateBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"name":"Changed","code":"CH","level":3}`, id)
	},
}

func TestMRole_CRUDSuite(t *testing.T) {
	runCRUDSuite(t, roleSpec)
}
