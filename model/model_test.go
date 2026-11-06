package model

import "testing"

func TestTableName(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"MNotes", MNotes{}.TableName(), "m_notes"},
		{"MUser", MUser{}.TableName(), "m_user"},
		{"MBiodata", MBiodata{}.TableName(), "m_biodata"},
		{"MRole", MRole{}.TableName(), "m_role"},
		{"TToken", TToken{}.TableName(), "t_token"},
		{"TResetPassword", TResetPassword{}.TableName(), "t_reset_password"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}
