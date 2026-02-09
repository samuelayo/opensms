package users

import "time"

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=12"`
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	MiddleName string `json:"middle_name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Role       string `json:"role" validate:"required"`
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	SchoolID   string `json:"school_id,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	MiddleName string `json:"middle_name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Status     string `json:"status,omitempty"`
}

// CreateStudentRequest represents a request to create a student
type CreateStudentRequest struct {
	// User fields
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=12"`
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	MiddleName string `json:"middle_name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	SchoolID   string `json:"school_id" validate:"required,uuid"`

	// Student fields
	AdmissionNumber         string `json:"admission_number" validate:"required"`
	AdmissionDate           string `json:"admission_date" validate:"required"`
	DateOfBirth             string `json:"date_of_birth" validate:"required"`
	Gender                  string `json:"gender" validate:"required,oneof=male female other prefer_not_to_say"`
	BloodGroup              string `json:"blood_group,omitempty"`
	Nationality             string `json:"nationality,omitempty"`
	Religion                string `json:"religion,omitempty"`
	Address                 string `json:"address,omitempty"`
	City                    string `json:"city,omitempty"`
	State                   string `json:"state,omitempty"`
	PostalCode              string `json:"postal_code,omitempty"`
	EmergencyContactName    string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone   string `json:"emergency_contact_phone,omitempty"`
	EmergencyContactRelation string `json:"emergency_contact_relation,omitempty"`
	MedicalConditions       string `json:"medical_conditions,omitempty"`
	Allergies               string `json:"allergies,omitempty"`
	PreviousSchool          string `json:"previous_school,omitempty"`
	CurrentGradeLevelID     string `json:"current_grade_level_id,omitempty"`
	CurrentClassID          string `json:"current_class_id,omitempty"`
}

// UpdateStudentRequest represents a request to update a student
type UpdateStudentRequest struct {
	Address                 string `json:"address,omitempty"`
	City                    string `json:"city,omitempty"`
	State                   string `json:"state,omitempty"`
	PostalCode              string `json:"postal_code,omitempty"`
	EmergencyContactName    string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone   string `json:"emergency_contact_phone,omitempty"`
	EmergencyContactRelation string `json:"emergency_contact_relation,omitempty"`
	MedicalConditions       string `json:"medical_conditions,omitempty"`
	Allergies               string `json:"allergies,omitempty"`
	CurrentGradeLevelID     string `json:"current_grade_level_id,omitempty"`
	CurrentClassID          string `json:"current_class_id,omitempty"`
	IsActive                *bool  `json:"is_active,omitempty"`
}

// CreateTeacherRequest represents a request to create a teacher
type CreateTeacherRequest struct {
	// User fields
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=12"`
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	MiddleName string `json:"middle_name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	SchoolID   string `json:"school_id" validate:"required,uuid"`

	// Teacher fields
	EmployeeID      string  `json:"employee_id" validate:"required"`
	DepartmentID    string  `json:"department_id,omitempty"`
	DateOfBirth     string  `json:"date_of_birth" validate:"required"`
	Gender          string  `json:"gender" validate:"required,oneof=male female other prefer_not_to_say"`
	DateOfJoining   string  `json:"date_of_joining" validate:"required"`
	Qualification   string  `json:"qualification,omitempty"`
	Specialization  string  `json:"specialization,omitempty"`
	ExperienceYears int     `json:"experience_years,omitempty"`
	EmploymentType  string  `json:"employment_type,omitempty"`
	Salary          float64 `json:"salary,omitempty"`
}

// UpdateTeacherRequest represents a request to update a teacher
type UpdateTeacherRequest struct {
	DepartmentID    string  `json:"department_id,omitempty"`
	Qualification   string  `json:"qualification,omitempty"`
	Specialization  string  `json:"specialization,omitempty"`
	ExperienceYears int     `json:"experience_years,omitempty"`
	EmploymentType  string  `json:"employment_type,omitempty"`
	Salary          float64 `json:"salary,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
}

// UserDTO represents a user in API responses
type UserDTO struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenant_id"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	Status           string `json:"status"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	MiddleName       string `json:"middle_name,omitempty"`
	Phone            string `json:"phone,omitempty"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	EmailVerified    bool   `json:"email_verified"`
	TwoFactorEnabled bool   `json:"two_factor_enabled"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// StudentDTO represents a student in API responses
type StudentDTO struct {
	ID                      string  `json:"id"`
	TenantID                string  `json:"tenant_id"`
	UserID                  string  `json:"user_id"`
	SchoolID                string  `json:"school_id"`
	AdmissionNumber         string  `json:"admission_number"`
	AdmissionDate           string  `json:"admission_date"`
	DateOfBirth             string  `json:"date_of_birth"`
	Gender                  string  `json:"gender"`
	BloodGroup              string  `json:"blood_group,omitempty"`
	Nationality             string  `json:"nationality,omitempty"`
	Religion                string  `json:"religion,omitempty"`
	Address                 string  `json:"address,omitempty"`
	City                    string  `json:"city,omitempty"`
	State                   string  `json:"state,omitempty"`
	PostalCode              string  `json:"postal_code,omitempty"`
	EmergencyContactName    string  `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone   string  `json:"emergency_contact_phone,omitempty"`
	EmergencyContactRelation string  `json:"emergency_contact_relation,omitempty"`
	MedicalConditions       string  `json:"medical_conditions,omitempty"`
	Allergies               string  `json:"allergies,omitempty"`
	PreviousSchool          string  `json:"previous_school,omitempty"`
	CurrentGradeLevelID     string  `json:"current_grade_level_id,omitempty"`
	CurrentClassID          string  `json:"current_class_id,omitempty"`
	IsActive                bool    `json:"is_active"`
	User                    *UserDTO `json:"user,omitempty"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
}

// TeacherDTO represents a teacher in API responses
type TeacherDTO struct {
	ID              string   `json:"id"`
	TenantID        string   `json:"tenant_id"`
	UserID          string   `json:"user_id"`
	SchoolID        string   `json:"school_id"`
	EmployeeID      string   `json:"employee_id"`
	DepartmentID    string   `json:"department_id,omitempty"`
	DateOfBirth     string   `json:"date_of_birth"`
	Gender          string   `json:"gender"`
	DateOfJoining   string   `json:"date_of_joining"`
	Qualification   string   `json:"qualification,omitempty"`
	Specialization  string   `json:"specialization,omitempty"`
	ExperienceYears int      `json:"experience_years"`
	EmploymentType  string   `json:"employment_type,omitempty"`
	Salary          float64  `json:"salary,omitempty"`
	IsActive        bool     `json:"is_active"`
	User            *UserDTO `json:"user,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// ToUserDTO converts a User to UserDTO
func ToUserDTO(user *User) *UserDTO {
	dto := &UserDTO{
		ID:               user.ID,
		TenantID:         user.TenantID,
		Email:            user.Email,
		Role:             user.Role,
		Status:           user.Status,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		CreatedAt:        user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        user.UpdatedAt.Format(time.RFC3339),
	}

	if user.MiddleName != nil {
		dto.MiddleName = *user.MiddleName
	}
	if user.Phone != nil {
		dto.Phone = *user.Phone
	}
	if user.AvatarURL != nil {
		dto.AvatarURL = *user.AvatarURL
	}

	return dto
}

// ToStudentDTO converts a Student to StudentDTO
func ToStudentDTO(student *Student, user *User) *StudentDTO {
	dto := &StudentDTO{
		ID:               student.ID,
		TenantID:         student.TenantID,
		UserID:           student.UserID,
		SchoolID:         student.SchoolID,
		AdmissionNumber:  student.AdmissionNumber,
		AdmissionDate:    student.AdmissionDate.Format("2006-01-02"),
		DateOfBirth:      student.DateOfBirth.Format("2006-01-02"),
		Gender:           student.Gender,
		IsActive:         student.IsActive,
		CreatedAt:        student.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        student.UpdatedAt.Format(time.RFC3339),
	}

	if student.BloodGroup != nil {
		dto.BloodGroup = *student.BloodGroup
	}
	if student.Nationality != nil {
		dto.Nationality = *student.Nationality
	}
	if student.Religion != nil {
		dto.Religion = *student.Religion
	}
	if student.Address != nil {
		dto.Address = *student.Address
	}
	if student.City != nil {
		dto.City = *student.City
	}
	if student.State != nil {
		dto.State = *student.State
	}
	if student.PostalCode != nil {
		dto.PostalCode = *student.PostalCode
	}
	if student.EmergencyContactName != nil {
		dto.EmergencyContactName = *student.EmergencyContactName
	}
	if student.EmergencyContactPhone != nil {
		dto.EmergencyContactPhone = *student.EmergencyContactPhone
	}
	if student.EmergencyContactRelation != nil {
		dto.EmergencyContactRelation = *student.EmergencyContactRelation
	}
	if student.MedicalConditions != nil {
		dto.MedicalConditions = *student.MedicalConditions
	}
	if student.Allergies != nil {
		dto.Allergies = *student.Allergies
	}
	if student.PreviousSchool != nil {
		dto.PreviousSchool = *student.PreviousSchool
	}
	if student.CurrentGradeLevelID != nil {
		dto.CurrentGradeLevelID = *student.CurrentGradeLevelID
	}
	if student.CurrentClassID != nil {
		dto.CurrentClassID = *student.CurrentClassID
	}

	if user != nil {
		dto.User = ToUserDTO(user)
	}

	return dto
}

// ToTeacherDTO converts a Teacher to TeacherDTO
func ToTeacherDTO(teacher *Teacher, user *User) *TeacherDTO {
	dto := &TeacherDTO{
		ID:              teacher.ID,
		TenantID:        teacher.TenantID,
		UserID:          teacher.UserID,
		SchoolID:        teacher.SchoolID,
		EmployeeID:      teacher.EmployeeID,
		DateOfBirth:     teacher.DateOfBirth.Format("2006-01-02"),
		Gender:          teacher.Gender,
		DateOfJoining:   teacher.DateOfJoining.Format("2006-01-02"),
		ExperienceYears: teacher.ExperienceYears,
		IsActive:        teacher.IsActive,
		CreatedAt:       teacher.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       teacher.UpdatedAt.Format(time.RFC3339),
	}

	if teacher.DepartmentID != nil {
		dto.DepartmentID = *teacher.DepartmentID
	}
	if teacher.Qualification != nil {
		dto.Qualification = *teacher.Qualification
	}
	if teacher.Specialization != nil {
		dto.Specialization = *teacher.Specialization
	}
	if teacher.EmploymentType != nil {
		dto.EmploymentType = *teacher.EmploymentType
	}
	if teacher.Salary != nil {
		dto.Salary = *teacher.Salary
	}

	if user != nil {
		dto.User = ToUserDTO(user)
	}

	return dto
}
