package users

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrStudentNotFound = errors.New("student not found")
	ErrTeacherNotFound = errors.New("teacher not found")
)

// User represents a user in the system
type User struct {
	ID               string
	TenantID         string
	Email            string
	Role             string
	Status           string
	FirstName        string
	LastName         string
	MiddleName       *string
	Phone            *string
	AvatarURL        *string
	EmailVerified    bool
	TwoFactorEnabled bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Student represents a student
type Student struct {
	ID                      string
	TenantID                string
	UserID                  string
	SchoolID                string
	AdmissionNumber         string
	AdmissionDate           time.Time
	DateOfBirth             time.Time
	Gender                  string
	BloodGroup              *string
	Nationality             *string
	Religion                *string
	Address                 *string
	City                    *string
	State                   *string
	PostalCode              *string
	EmergencyContactName    *string
	EmergencyContactPhone   *string
	EmergencyContactRelation *string
	MedicalConditions       *string
	Allergies               *string
	PreviousSchool          *string
	CurrentGradeLevelID     *string
	CurrentClassID          *string
	IsActive                bool
	WithdrawalDate          *time.Time
	WithdrawalReason        *string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// Teacher represents a teacher
type Teacher struct {
	ID              string
	TenantID        string
	UserID          string
	SchoolID        string
	EmployeeID      string
	DepartmentID    *string
	DateOfBirth     time.Time
	Gender          string
	DateOfJoining   time.Time
	Qualification   *string
	Specialization  *string
	ExperienceYears int
	EmploymentType  *string
	Salary          *float64
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Repository defines the interface for user data access
type Repository interface {
	// User operations
	ListUsers(ctx context.Context, limit, offset int) ([]*User, int64, error)
	GetUser(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, user *User, passwordHash string) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error

	// Student operations
	ListStudents(ctx context.Context, limit, offset int) ([]*Student, int64, error)
	GetStudent(ctx context.Context, id string) (*Student, error)
	GetStudentByUserID(ctx context.Context, userID string) (*Student, error)
	CreateStudent(ctx context.Context, student *Student) error
	UpdateStudent(ctx context.Context, student *Student) error

	// Teacher operations
	ListTeachers(ctx context.Context, limit, offset int) ([]*Teacher, int64, error)
	GetTeacher(ctx context.Context, id string) (*Teacher, error)
	GetTeacherByUserID(ctx context.Context, userID string) (*Teacher, error)
	CreateTeacher(ctx context.Context, teacher *Teacher) error
	UpdateTeacher(ctx context.Context, teacher *Teacher) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ListUsers retrieves a paginated list of users
func (r *PostgresRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM users WHERE status != 'deleted'`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated users
	query := `
		SELECT id, tenant_id, email, role, status, first_name, last_name,
			middle_name, phone, avatar_url, email_verified, two_factor_enabled,
			created_at, updated_at
		FROM users
		WHERE status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.TenantID, &user.Email, &user.Role, &user.Status,
			&user.FirstName, &user.LastName, &user.MiddleName, &user.Phone,
			&user.AvatarURL, &user.EmailVerified, &user.TwoFactorEnabled,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetUser retrieves a user by ID
func (r *PostgresRepository) GetUser(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, tenant_id, email, role, status, first_name, last_name,
			middle_name, phone, avatar_url, email_verified, two_factor_enabled,
			created_at, updated_at
		FROM users
		WHERE id = $1 AND status != 'deleted'
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Role, &user.Status,
		&user.FirstName, &user.LastName, &user.MiddleName, &user.Phone,
		&user.AvatarURL, &user.EmailVerified, &user.TwoFactorEnabled,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// CreateUser creates a new user
func (r *PostgresRepository) CreateUser(ctx context.Context, user *User, passwordHash string) error {
	query := `
		INSERT INTO users (
			id, tenant_id, email, password_hash, role, status,
			first_name, last_name, middle_name, phone, avatar_url,
			email_verified, two_factor_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.TenantID, user.Email, passwordHash, user.Role, user.Status,
		user.FirstName, user.LastName, user.MiddleName, user.Phone, user.AvatarURL,
		user.EmailVerified, user.TwoFactorEnabled, time.Now(), time.Now(),
	)

	return err
}

// UpdateUser updates an existing user
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users SET
			first_name = $1, last_name = $2, middle_name = $3,
			phone = $4, avatar_url = $5, status = $6, updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.Exec(ctx, query,
		user.FirstName, user.LastName, user.MiddleName, user.Phone,
		user.AvatarURL, user.Status, time.Now(), user.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser soft deletes a user
func (r *PostgresRepository) DeleteUser(ctx context.Context, id string) error {
	query := `UPDATE users SET status = 'deleted', updated_at = $1 WHERE id = $2`
	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListStudents retrieves a paginated list of students
func (r *PostgresRepository) ListStudents(ctx context.Context, limit, offset int) ([]*Student, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM students WHERE is_active = true`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated students
	query := `
		SELECT id, tenant_id, user_id, school_id, admission_number, admission_date,
			date_of_birth, gender, blood_group, nationality, religion,
			address, city, state, postal_code,
			emergency_contact_name, emergency_contact_phone, emergency_contact_relationship,
			medical_conditions, allergies, previous_school,
			current_grade_level_id, current_class_id, is_active,
			withdrawal_date, withdrawal_reason, created_at, updated_at
		FROM students
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []*Student
	for rows.Next() {
		student := &Student{}
		err := rows.Scan(
			&student.ID, &student.TenantID, &student.UserID, &student.SchoolID,
			&student.AdmissionNumber, &student.AdmissionDate, &student.DateOfBirth,
			&student.Gender, &student.BloodGroup, &student.Nationality, &student.Religion,
			&student.Address, &student.City, &student.State, &student.PostalCode,
			&student.EmergencyContactName, &student.EmergencyContactPhone,
			&student.EmergencyContactRelation, &student.MedicalConditions,
			&student.Allergies, &student.PreviousSchool, &student.CurrentGradeLevelID,
			&student.CurrentClassID, &student.IsActive, &student.WithdrawalDate,
			&student.WithdrawalReason, &student.CreatedAt, &student.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		students = append(students, student)
	}

	return students, total, nil
}

// GetStudent retrieves a student by ID
func (r *PostgresRepository) GetStudent(ctx context.Context, id string) (*Student, error) {
	query := `
		SELECT id, tenant_id, user_id, school_id, admission_number, admission_date,
			date_of_birth, gender, blood_group, nationality, religion,
			address, city, state, postal_code,
			emergency_contact_name, emergency_contact_phone, emergency_contact_relationship,
			medical_conditions, allergies, previous_school,
			current_grade_level_id, current_class_id, is_active,
			withdrawal_date, withdrawal_reason, created_at, updated_at
		FROM students
		WHERE id = $1
	`

	student := &Student{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&student.ID, &student.TenantID, &student.UserID, &student.SchoolID,
		&student.AdmissionNumber, &student.AdmissionDate, &student.DateOfBirth,
		&student.Gender, &student.BloodGroup, &student.Nationality, &student.Religion,
		&student.Address, &student.City, &student.State, &student.PostalCode,
		&student.EmergencyContactName, &student.EmergencyContactPhone,
		&student.EmergencyContactRelation, &student.MedicalConditions,
		&student.Allergies, &student.PreviousSchool, &student.CurrentGradeLevelID,
		&student.CurrentClassID, &student.IsActive, &student.WithdrawalDate,
		&student.WithdrawalReason, &student.CreatedAt, &student.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return student, nil
}

// GetStudentByUserID retrieves a student by user ID
func (r *PostgresRepository) GetStudentByUserID(ctx context.Context, userID string) (*Student, error) {
	query := `
		SELECT id, tenant_id, user_id, school_id, admission_number, admission_date,
			date_of_birth, gender, blood_group, nationality, religion,
			address, city, state, postal_code,
			emergency_contact_name, emergency_contact_phone, emergency_contact_relationship,
			medical_conditions, allergies, previous_school,
			current_grade_level_id, current_class_id, is_active,
			withdrawal_date, withdrawal_reason, created_at, updated_at
		FROM students
		WHERE user_id = $1
	`

	student := &Student{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&student.ID, &student.TenantID, &student.UserID, &student.SchoolID,
		&student.AdmissionNumber, &student.AdmissionDate, &student.DateOfBirth,
		&student.Gender, &student.BloodGroup, &student.Nationality, &student.Religion,
		&student.Address, &student.City, &student.State, &student.PostalCode,
		&student.EmergencyContactName, &student.EmergencyContactPhone,
		&student.EmergencyContactRelation, &student.MedicalConditions,
		&student.Allergies, &student.PreviousSchool, &student.CurrentGradeLevelID,
		&student.CurrentClassID, &student.IsActive, &student.WithdrawalDate,
		&student.WithdrawalReason, &student.CreatedAt, &student.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return student, nil
}

// CreateStudent creates a new student
func (r *PostgresRepository) CreateStudent(ctx context.Context, student *Student) error {
	query := `
		INSERT INTO students (
			id, tenant_id, user_id, school_id, admission_number, admission_date,
			date_of_birth, gender, blood_group, nationality, religion,
			address, city, state, postal_code,
			emergency_contact_name, emergency_contact_phone, emergency_contact_relationship,
			medical_conditions, allergies, previous_school,
			current_grade_level_id, current_class_id, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
	`

	_, err := r.db.Exec(ctx, query,
		student.ID, student.TenantID, student.UserID, student.SchoolID,
		student.AdmissionNumber, student.AdmissionDate, student.DateOfBirth,
		student.Gender, student.BloodGroup, student.Nationality, student.Religion,
		student.Address, student.City, student.State, student.PostalCode,
		student.EmergencyContactName, student.EmergencyContactPhone,
		student.EmergencyContactRelation, student.MedicalConditions,
		student.Allergies, student.PreviousSchool, student.CurrentGradeLevelID,
		student.CurrentClassID, student.IsActive, time.Now(), time.Now(),
	)

	return err
}

// UpdateStudent updates an existing student
func (r *PostgresRepository) UpdateStudent(ctx context.Context, student *Student) error {
	query := `
		UPDATE students SET
			address = $1, city = $2, state = $3, postal_code = $4,
			emergency_contact_name = $5, emergency_contact_phone = $6,
			emergency_contact_relationship = $7, medical_conditions = $8,
			allergies = $9, current_grade_level_id = $10, current_class_id = $11,
			is_active = $12, updated_at = $13
		WHERE id = $14
	`

	result, err := r.db.Exec(ctx, query,
		student.Address, student.City, student.State, student.PostalCode,
		student.EmergencyContactName, student.EmergencyContactPhone,
		student.EmergencyContactRelation, student.MedicalConditions,
		student.Allergies, student.CurrentGradeLevelID, student.CurrentClassID,
		student.IsActive, time.Now(), student.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrStudentNotFound
	}

	return nil
}

// ListTeachers retrieves a paginated list of teachers
func (r *PostgresRepository) ListTeachers(ctx context.Context, limit, offset int) ([]*Teacher, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM teachers WHERE is_active = true`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated teachers
	query := `
		SELECT id, tenant_id, user_id, school_id, employee_id, department_id,
			date_of_birth, gender, date_of_joining, qualification, specialization,
			experience_years, employment_type, salary, is_active, created_at, updated_at
		FROM teachers
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var teachers []*Teacher
	for rows.Next() {
		teacher := &Teacher{}
		err := rows.Scan(
			&teacher.ID, &teacher.TenantID, &teacher.UserID, &teacher.SchoolID,
			&teacher.EmployeeID, &teacher.DepartmentID, &teacher.DateOfBirth,
			&teacher.Gender, &teacher.DateOfJoining, &teacher.Qualification,
			&teacher.Specialization, &teacher.ExperienceYears, &teacher.EmploymentType,
			&teacher.Salary, &teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		teachers = append(teachers, teacher)
	}

	return teachers, total, nil
}

// GetTeacher retrieves a teacher by ID
func (r *PostgresRepository) GetTeacher(ctx context.Context, id string) (*Teacher, error) {
	query := `
		SELECT id, tenant_id, user_id, school_id, employee_id, department_id,
			date_of_birth, gender, date_of_joining, qualification, specialization,
			experience_years, employment_type, salary, is_active, created_at, updated_at
		FROM teachers
		WHERE id = $1
	`

	teacher := &Teacher{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&teacher.ID, &teacher.TenantID, &teacher.UserID, &teacher.SchoolID,
		&teacher.EmployeeID, &teacher.DepartmentID, &teacher.DateOfBirth,
		&teacher.Gender, &teacher.DateOfJoining, &teacher.Qualification,
		&teacher.Specialization, &teacher.ExperienceYears, &teacher.EmploymentType,
		&teacher.Salary, &teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTeacherNotFound
		}
		return nil, err
	}

	return teacher, nil
}

// GetTeacherByUserID retrieves a teacher by user ID
func (r *PostgresRepository) GetTeacherByUserID(ctx context.Context, userID string) (*Teacher, error) {
	query := `
		SELECT id, tenant_id, user_id, school_id, employee_id, department_id,
			date_of_birth, gender, date_of_joining, qualification, specialization,
			experience_years, employment_type, salary, is_active, created_at, updated_at
		FROM teachers
		WHERE user_id = $1
	`

	teacher := &Teacher{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&teacher.ID, &teacher.TenantID, &teacher.UserID, &teacher.SchoolID,
		&teacher.EmployeeID, &teacher.DepartmentID, &teacher.DateOfBirth,
		&teacher.Gender, &teacher.DateOfJoining, &teacher.Qualification,
		&teacher.Specialization, &teacher.ExperienceYears, &teacher.EmploymentType,
		&teacher.Salary, &teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTeacherNotFound
		}
		return nil, err
	}

	return teacher, nil
}

// CreateTeacher creates a new teacher
func (r *PostgresRepository) CreateTeacher(ctx context.Context, teacher *Teacher) error {
	query := `
		INSERT INTO teachers (
			id, tenant_id, user_id, school_id, employee_id, department_id,
			date_of_birth, gender, date_of_joining, qualification, specialization,
			experience_years, employment_type, salary, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.Exec(ctx, query,
		teacher.ID, teacher.TenantID, teacher.UserID, teacher.SchoolID,
		teacher.EmployeeID, teacher.DepartmentID, teacher.DateOfBirth,
		teacher.Gender, teacher.DateOfJoining, teacher.Qualification,
		teacher.Specialization, teacher.ExperienceYears, teacher.EmploymentType,
		teacher.Salary, teacher.IsActive, time.Now(), time.Now(),
	)

	return err
}

// UpdateTeacher updates an existing teacher
func (r *PostgresRepository) UpdateTeacher(ctx context.Context, teacher *Teacher) error {
	query := `
		UPDATE teachers SET
			department_id = $1, qualification = $2, specialization = $3,
			experience_years = $4, employment_type = $5, salary = $6,
			is_active = $7, updated_at = $8
		WHERE id = $9
	`

	result, err := r.db.Exec(ctx, query,
		teacher.DepartmentID, teacher.Qualification, teacher.Specialization,
		teacher.ExperienceYears, teacher.EmploymentType, teacher.Salary,
		teacher.IsActive, time.Now(), teacher.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrTeacherNotFound
	}

	return nil
}
