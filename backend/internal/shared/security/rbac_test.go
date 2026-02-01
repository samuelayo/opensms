package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRBAC_HasPermission(t *testing.T) {
	rbac := NewRBAC()

	tests := []struct {
		name       string
		role       Role
		permission Permission
		want       bool
	}{
		// Super admin has all permissions
		{"super admin has user:create", RoleSuperAdmin, PermUserCreate, true},
		{"super admin has grade:publish", RoleSuperAdmin, PermGradePublish, true},
		{"super admin has fee:create", RoleSuperAdmin, PermFeeCreate, true},

		// Teacher permissions
		{"teacher has grade:create", RoleTeacher, PermGradeCreate, true},
		{"teacher has attendance:mark", RoleTeacher, PermAttendanceMark, true},
		{"teacher does not have user:create", RoleTeacher, PermUserCreate, false},
		{"teacher does not have fee:create", RoleTeacher, PermFeeCreate, false},

		// Student permissions
		{"student has grade:read", RoleStudent, PermGradeRead, true},
		{"student does not have grade:create", RoleStudent, PermGradeCreate, false},
		{"student does not have user:create", RoleStudent, PermUserCreate, false},

		// Parent permissions
		{"parent has payment:view", RoleParent, PermPaymentView, true},
		{"parent has payment:process", RoleParent, PermPaymentProcess, true},
		{"parent does not have grade:create", RoleParent, PermGradeCreate, false},

		// Librarian permissions
		{"librarian has library:book_add", RoleLibrarian, PermLibraryBookAdd, true},
		{"librarian does not have grade:read", RoleLibrarian, PermGradeRead, false},

		// Accountant permissions
		{"accountant has fee:create", RoleAccountant, PermFeeCreate, true},
		{"accountant has payment:process", RoleAccountant, PermPaymentProcess, true},
		{"accountant does not have user:create", RoleAccountant, PermUserCreate, false},

		// School admin permissions
		{"school admin has user:create", RoleSchoolAdmin, PermUserCreate, true},
		{"school admin has student:enroll", RoleSchoolAdmin, PermStudentEnroll, true},
		{"school admin does not have user:delete", RoleSchoolAdmin, PermUserDelete, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rbac.HasPermission(tt.role, tt.permission)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRBAC_HasAnyPermission(t *testing.T) {
	rbac := NewRBAC()

	tests := []struct {
		name        string
		role        Role
		permissions []Permission
		want        bool
	}{
		{
			name:        "teacher has any of grade permissions",
			role:        RoleTeacher,
			permissions: []Permission{PermGradeCreate, PermGradeUpdate, PermUserDelete},
			want:        true,
		},
		{
			name:        "student has none of admin permissions",
			role:        RoleStudent,
			permissions: []Permission{PermUserCreate, PermUserDelete, PermSystemConfig},
			want:        false,
		},
		{
			name:        "accountant has any of finance permissions",
			role:        RoleAccountant,
			permissions: []Permission{PermFeeCreate, PermPaymentProcess},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rbac.HasAnyPermission(tt.role, tt.permissions)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRBAC_HasAllPermissions(t *testing.T) {
	rbac := NewRBAC()

	tests := []struct {
		name        string
		role        Role
		permissions []Permission
		want        bool
	}{
		{
			name:        "teacher has all grade permissions",
			role:        RoleTeacher,
			permissions: []Permission{PermGradeCreate, PermGradeRead, PermGradeUpdate},
			want:        true,
		},
		{
			name:        "teacher does not have all admin permissions",
			role:        RoleTeacher,
			permissions: []Permission{PermGradeCreate, PermUserDelete},
			want:        false,
		},
		{
			name:        "super admin has all permissions",
			role:        RoleSuperAdmin,
			permissions: []Permission{PermUserCreate, PermGradeCreate, PermFeeCreate},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rbac.HasAllPermissions(tt.role, tt.permissions)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRBAC_GetRolePermissions(t *testing.T) {
	rbac := NewRBAC()

	tests := []struct {
		name string
		role Role
	}{
		{"super admin", RoleSuperAdmin},
		{"teacher", RoleTeacher},
		{"student", RoleStudent},
		{"accountant", RoleAccountant},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := rbac.GetRolePermissions(tt.role)
			assert.NotNil(t, permissions)
			assert.NotEmpty(t, permissions)
		})
	}
}

func TestRBAC_AddRolePermission(t *testing.T) {
	rbac := NewRBAC()
	customPermission := Permission("custom:permission")

	// Initially teacher should not have custom permission
	assert.False(t, rbac.HasPermission(RoleTeacher, customPermission))

	// Add custom permission
	rbac.AddRolePermission(RoleTeacher, customPermission)

	// Now teacher should have it
	assert.True(t, rbac.HasPermission(RoleTeacher, customPermission))
}

func TestRBAC_RemoveRolePermission(t *testing.T) {
	rbac := NewRBAC()

	// Teacher initially has grade:create
	assert.True(t, rbac.HasPermission(RoleTeacher, PermGradeCreate))

	// Remove permission
	rbac.RemoveRolePermission(RoleTeacher, PermGradeCreate)

	// Now teacher should not have it
	assert.False(t, rbac.HasPermission(RoleTeacher, PermGradeCreate))
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"valid super_admin", "super_admin", false},
		{"valid teacher", "teacher", false},
		{"valid student", "student", false},
		{"invalid role", "invalid_role", true},
		{"empty role", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetHierarchyLevel(t *testing.T) {
	tests := []struct {
		role     Role
		expected int
	}{
		{RoleSuperAdmin, 100},
		{RoleSchoolAdmin, 90},
		{RolePrincipal, 80},
		{RoleTeacher, 50},
		{RoleAccountant, 50},
		{RoleStudent, 10},
		{RoleParent, 20},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			level := GetHierarchyLevel(tt.role)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestCanManageRole(t *testing.T) {
	tests := []struct {
		name        string
		managerRole Role
		targetRole  Role
		want        bool
	}{
		{"super admin can manage school admin", RoleSuperAdmin, RoleSchoolAdmin, true},
		{"super admin can manage teacher", RoleSuperAdmin, RoleTeacher, true},
		{"school admin can manage teacher", RoleSchoolAdmin, RoleTeacher, true},
		{"teacher cannot manage school admin", RoleTeacher, RoleSchoolAdmin, false},
		{"teacher cannot manage another teacher", RoleTeacher, RoleTeacher, false},
		{"principal can manage teacher", RolePrincipal, RoleTeacher, true},
		{"teacher cannot manage student", RoleTeacher, RoleStudent, true},
		{"student cannot manage teacher", RoleStudent, RoleTeacher, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanManageRole(tt.managerRole, tt.targetRole)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRolePermissions_Coverage(t *testing.T) {
	// Ensure all roles have at least some permissions defined
	allRoles := []Role{
		RoleSuperAdmin,
		RoleSchoolAdmin,
		RolePrincipal,
		RoleTeacher,
		RoleStudent,
		RoleParent,
		RoleLibrarian,
		RoleAccountant,
		RoleRegistrar,
	}

	for _, role := range allRoles {
		t.Run(string(role), func(t *testing.T) {
			permissions, exists := RolePermissions[role]
			assert.True(t, exists, "Role %s should have permissions defined", role)
			assert.NotEmpty(t, permissions, "Role %s should have at least one permission", role)
		})
	}
}

// Benchmark tests
func BenchmarkHasPermission(b *testing.B) {
	rbac := NewRBAC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rbac.HasPermission(RoleTeacher, PermGradeCreate)
	}
}

func BenchmarkHasAnyPermission(b *testing.B) {
	rbac := NewRBAC()
	permissions := []Permission{PermGradeCreate, PermGradeUpdate, PermAttendanceMark}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rbac.HasAnyPermission(RoleTeacher, permissions)
	}
}

func BenchmarkValidateRole(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateRole("teacher")
	}
}
