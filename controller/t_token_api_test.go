package controller

import (
	"fmt"
	"testing"
)

var tokenSpec = crudSpec{
	base: "/t_token",
	createBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"email":"a@b.c","userId":1,"token":"tok%[1]d","usedFor":"reset",`+nestedUserJSON+`}`, id)
	},
	updateBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"email":"a@b.c","userId":1,"token":"new","usedFor":"reset",`+nestedUserJSON+`}`, id)
	},
}

func TestTToken_CRUDSuite(t *testing.T) {
	runCRUDSuite(t, tokenSpec)
}
