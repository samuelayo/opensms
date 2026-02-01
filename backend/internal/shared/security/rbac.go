package security

import (
	"fmt"
	"strings"
)

// Role represents a user role in the system
type Role string

const (
	RoleSuperAdmin     Role = "super_admin"      // Platform super administrator
	RoleSchoolAdmin    Role = "school_admin"     // School/institution administrator
	RolePrincipal      Role = "principal"        // School principal
	RoleTeacher        Role = "teacher"          // Teacher
	RoleStudent        Role = "student"          // Student
	RoleParent         Role = "parent"           // Parent/Guardian
	RoleLibrarian      Role = "librarian"        // Librarian
	RoleAccountant     Role = "accountant"       // Accountant/Finance officer
	RoleRegistrar      Role = "registrar"        // Registrar/Admissions officer
	RoleCounselor      Role = "counselor"        // School counselor
	RoleNurse          Role = "nurse"            // School nurse
	RoleTransportAdmin Role = "transport_admin"  // Transport administrator
	RoleHostelWarden   Role = "hostel_warden"    // Hostel warden
)

// Permission represents a specific permission in the system
type Permission string

// User Management Permissions
const (
	PermUserCreate        Permission = "user:create"
	PermUserRead          Permission = "user:read"
	PermUserUpdate        Permission = "user:update"
	PermUserDelete        Permission = "user:delete"
	PermUserManageRoles   Permission = "user:manage_roles"
	PermUserResetPassword Permission = "user:reset_password"
)

// Student Management Permissions
const (
	PermStudentCreate      Permission = "student:create"
	PermStudentRead        Permission = "student:read"
	PermStudentUpdate      Permission = "student:update"
	PermStudentDelete      Permission = "student:delete"
	PermStudentEnroll      Permission = "student:enroll"
	PermStudentWithdraw    Permission = "student:withdraw"
	PermStudentPromote     Permission = "student:promote"
	PermStudentViewRecords Permission = "student:view_records"
)

// Academic Permissions
const (
	PermGradeCreate     Permission = "grade:create"
	PermGradeRead       Permission = "grade:read"
	PermGradeUpdate     Permission = "grade:update"
	PermGradeDelete     Permission = "grade:delete"
	PermGradePublish    Permission = "grade:publish"
	PermAttendanceMark  Permission = "attendance:mark"
	PermAttendanceRead  Permission = "attendance:read"
	PermAssignmentCreate Permission = "assignment:create"
	PermAssignmentGrade  Permission = "assignment:grade"
	PermExamCreate      Permission = "exam:create"
	PermExamManage      Permission = "exam:manage"
)

// Finance Permissions
const (
	PermFeeCreate       Permission = "fee:create"
	PermFeeRead         Permission = "fee:read"
	PermFeeUpdate       Permission = "fee:update"
	PermFeeDelete       Permission = "fee:delete"
	PermPaymentProcess  Permission = "payment:process"
	PermPaymentView     Permission = "payment:view"
	PermInvoiceGenerate Permission = "invoice:generate"
	PermInvoiceView     Permission = "invoice:view"
	PermReportFinancial Permission = "report:financial"
)

// Administration Permissions
const (
	PermSchoolCreate    Permission = "school:create"
	PermSchoolUpdate    Permission = "school:update"
	PermClassCreate     Permission = "class:create"
	PermClassManage     Permission = "class:manage"
	PermTimetableCreate Permission = "timetable:create"
	PermTimetableManage Permission = "timetable:manage"
)

// Library Permissions
const (
	PermLibraryBookAdd    Permission = "library:book_add"
	PermLibraryBookIssue  Permission = "library:book_issue"
	PermLibraryBookReturn Permission = "library:book_return"
	PermLibraryManage     Permission = "library:manage"
)

// Communication Permissions
const (
	PermNotificationSend Permission = "notification:send"
	PermAnnouncementPost Permission = "announcement:post"
	PermMessageSend      Permission = "message:send"
)

// System Permissions
const (
	PermSystemConfig      Permission = "system:config"
	PermSystemBackup      Permission = "system:backup"
	PermSystemAuditLog    Permission = "system:audit_log"
	PermReportGenerate    Permission = "report:generate"
	PermReportViewAll     Permission = "report:view_all"
)

// RolePermissions maps roles to their default permissions
var RolePermissions = map[Role][]Permission{
	RoleSuperAdmin: {
		// Super admin has all permissions
		PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete, PermUserManageRoles, PermUserResetPassword,
		PermStudentCreate, PermStudentRead, PermStudentUpdate, PermStudentDelete, PermStudentEnroll, PermStudentWithdraw, PermStudentPromote, PermStudentViewRecords,
		PermGradeCreate, PermGradeRead, PermGradeUpdate, PermGradeDelete, PermGradePublish,
		PermAttendanceMark, PermAttendanceRead,
		PermAssignmentCreate, PermAssignmentGrade,
		PermExamCreate, PermExamManage,
		PermFeeCreate, PermFeeRead, PermFeeUpdate, PermFeeDelete,
		PermPaymentProcess, PermPaymentView,
		PermInvoiceGenerate, PermInvoiceView,
		PermReportFinancial,
		PermSchoolCreate, PermSchoolUpdate,
		PermClassCreate, PermClassManage,
		PermTimetableCreate, PermTimetableManage,
		PermLibraryBookAdd, PermLibraryBookIssue, PermLibraryBookReturn, PermLibraryManage,
		PermNotificationSend, PermAnnouncementPost, PermMessageSend,
		PermSystemConfig, PermSystemBackup, PermSystemAuditLog,
		PermReportGenerate, PermReportViewAll,
	},
	RoleSchoolAdmin: {
		PermUserCreate, PermUserRead, PermUserUpdate, PermUserResetPassword,
		PermStudentCreate, PermStudentRead, PermStudentUpdate, PermStudentEnroll, PermStudentWithdraw, PermStudentPromote, PermStudentViewRecords,
		PermGradeRead, PermGradePublish,
		PermAttendanceRead,
		PermExamManage,
		PermFeeCreate, PermFeeRead, PermFeeUpdate,
		PermPaymentView, PermInvoiceView,
		PermSchoolUpdate,
		PermClassCreate, PermClassManage,
		PermTimetableCreate, PermTimetableManage,
		PermNotificationSend, PermAnnouncementPost,
		PermReportGenerate, PermReportViewAll,
	},
	RolePrincipal: {
		PermUserRead, PermUserResetPassword,
		PermStudentRead, PermStudentViewRecords,
		PermGradeRead,
		PermAttendanceRead,
		PermFeeRead,
		PermPaymentView, PermInvoiceView,
		PermTimetableManage,
		PermAnnouncementPost,
		PermReportViewAll,
	},
	RoleTeacher: {
		PermStudentRead,
		PermGradeCreate, PermGradeRead, PermGradeUpdate,
		PermAttendanceMark, PermAttendanceRead,
		PermAssignmentCreate, PermAssignmentGrade,
		PermMessageSend,
	},
	RoleStudent: {
		PermGradeRead,
		PermAttendanceRead,
		PermAssignmentCreate, // Submit assignments
		PermPaymentView,
		PermMessageSend,
	},
	RoleParent: {
		PermStudentRead, // Read own children's data
		PermGradeRead,
		PermAttendanceRead,
		PermPaymentView, PermPaymentProcess,
		PermMessageSend,
	},
	RoleLibrarian: {
		PermLibraryBookAdd, PermLibraryBookIssue, PermLibraryBookReturn, PermLibraryManage,
	},
	RoleAccountant: {
		PermFeeCreate, PermFeeRead, PermFeeUpdate,
		PermPaymentProcess, PermPaymentView,
		PermInvoiceGenerate, PermInvoiceView,
		PermReportFinancial,
	},
	RoleRegistrar: {
		PermStudentCreate, PermStudentRead, PermStudentUpdate, PermStudentEnroll,
		PermUserCreate, PermUserRead,
	},
}

// RBAC handles role-based access control
type RBAC struct {
	rolePermissions map[Role][]Permission
}

// NewRBAC creates a new RBAC instance
func NewRBAC() *RBAC {
	return &RBAC{
		rolePermissions: RolePermissions,
	}
}

// HasPermission checks if a role has a specific permission
func (r *RBAC) HasPermission(role Role, permission Permission) bool {
	permissions, exists := r.rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
		// Support wildcard permissions (e.g., "user:*" grants all user permissions)
		if strings.HasSuffix(string(p), ":*") {
			prefix := strings.TrimSuffix(string(p), "*")
			if strings.HasPrefix(string(permission), prefix) {
				return true
			}
		}
	}

	return false
}

// HasAnyPermission checks if a role has any of the specified permissions
func (r *RBAC) HasAnyPermission(role Role, permissions []Permission) bool {
	for _, perm := range permissions {
		if r.HasPermission(role, perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the specified permissions
func (r *RBAC) HasAllPermissions(role Role, permissions []Permission) bool {
	for _, perm := range permissions {
		if !r.HasPermission(role, perm) {
			return false
		}
	}
	return true
}

// GetRolePermissions returns all permissions for a role
func (r *RBAC) GetRolePermissions(role Role) []Permission {
	return r.rolePermissions[role]
}

// AddRolePermission adds a permission to a role (for custom permissions)
func (r *RBAC) AddRolePermission(role Role, permission Permission) {
	if _, exists := r.rolePermissions[role]; !exists {
		r.rolePermissions[role] = []Permission{}
	}
	r.rolePermissions[role] = append(r.rolePermissions[role], permission)
}

// RemoveRolePermission removes a permission from a role
func (r *RBAC) RemoveRolePermission(role Role, permission Permission) {
	permissions, exists := r.rolePermissions[role]
	if !exists {
		return
	}

	var newPermissions []Permission
	for _, p := range permissions {
		if p != permission {
			newPermissions = append(newPermissions, p)
		}
	}
	r.rolePermissions[role] = newPermissions
}

// ValidateRole checks if a role is valid
func ValidateRole(role string) error {
	validRoles := []Role{
		RoleSuperAdmin, RoleSchoolAdmin, RolePrincipal, RoleTeacher,
		RoleStudent, RoleParent, RoleLibrarian, RoleAccountant,
		RoleRegistrar, RoleCounselor, RoleNurse, RoleTransportAdmin,
		RoleHostelWarden,
	}

	for _, validRole := range validRoles {
		if Role(role) == validRole {
			return nil
		}
	}

	return fmt.Errorf("invalid role: %s", role)
}

// GetHierarchyLevel returns the hierarchy level of a role (higher number = more authority)
func GetHierarchyLevel(role Role) int {
	hierarchy := map[Role]int{
		RoleSuperAdmin:     100,
		RoleSchoolAdmin:    90,
		RolePrincipal:      80,
		RoleTeacher:        50,
		RoleAccountant:     50,
		RoleRegistrar:      50,
		RoleLibrarian:      40,
		RoleCounselor:      40,
		RoleNurse:          40,
		RoleTransportAdmin: 40,
		RoleHostelWarden:   40,
		RoleParent:         20,
		RoleStudent:        10,
	}

	return hierarchy[role]
}

// CanManageRole checks if a role can manage another role
func CanManageRole(managerRole, targetRole Role) bool {
	return GetHierarchyLevel(managerRole) > GetHierarchyLevel(targetRole)
}
