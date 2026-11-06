package controller

import (
	"fmt"
	"testing"
)

var resetPasswordSpec = crudSpec{
	base: "/t_reset_password",
	createBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"oldPassword":"old%[1]d","newPassword":"new","resetFor":"user"}`, id)
	},
	updateBody: func(id int) string {
		return fmt.Sprintf(`{"id":%[1]d,"oldPassword":"o","newPassword":"n","resetFor":"user"}`, id)
	},
}

func TestTResetPassword_CRUDSuite(t *testing.T) {
	runCRUDSuite(t, resetPasswordSpec)
}
