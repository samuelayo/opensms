package academic

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrGradeNotFound          = errors.New("grade not found")
	ErrAttendanceNotFound     = errors.New("attendance not found")
	ErrClassNotFound          = errors.New("class not found")
	ErrSubjectNotFound        = errors.New("subject not found")
	ErrAcademicYearNotFound   = errors.New("academic year not found")
	ErrAssignmentNotFound     = errors.New("assignment not found")
)

// Grade represents a student grade
type Grade struct {
	ID             string
	TenantID       string
	StudentID      string
	ClassID        string
	SubjectID      string
	TermID         string
	AssessmentType string
	AssessmentName *string
	Score          float64
	MaxScore       float64
	Grade          *string
	GradePoint     *float64
	Remarks        *string
	GradedBy       *string
	GradedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Attendance represents a student attendance record
type Attendance struct {
	ID        string
	TenantID  string
	StudentID string
	ClassID   string
	Date      time.Time
	Status    string
	Remarks   *string
	MarkedBy  *string
	MarkedAt  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AcademicYear represents an academic year
type AcademicYear struct {
	ID        string
	TenantID  string
	SchoolID  *string
	Name      string
	StartDate time.Time
	EndDate   time.Time
	IsCurrent bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Class represents a class/section
type Class struct {
	ID              string
	TenantID        string
	SchoolID        string
	GradeLevelID    string
	DepartmentID    *string
	Name            string
	Section         *string
	ClassTeacherID  *string
	Capacity        int
	RoomNumber      *string
	AcademicYearID  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Subject represents a subject/course
type Subject struct {
	ID            string
	TenantID      string
	SchoolID      string
	Name          string
	Code          string
	Description   *string
	DepartmentID  *string
	CreditHours   int
	IsElective    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Assignment represents a homework/project assignment
type Assignment struct {
	ID             string
	TenantID       string
	ClassID        string
	SubjectID      string
	TeacherID      string
	Title          string
	Description    *string
	AssignmentType string
	MaxScore       float64
	DueDate        time.Time
	AssignedDate   time.Time
	IsPublished    bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Repository defines the interface for academic data access
type Repository interface {
	// Academic year operations
	ListAcademicYears(ctx context.Context, limit, offset int) ([]*AcademicYear, int64, error)
	GetAcademicYear(ctx context.Context, id string) (*AcademicYear, error)
	CreateAcademicYear(ctx context.Context, year *AcademicYear) error

	// Class operations
	ListClasses(ctx context.Context, limit, offset int) ([]*Class, int64, error)
	GetClass(ctx context.Context, id string) (*Class, error)
	CreateClass(ctx context.Context, class *Class) error
	UpdateClass(ctx context.Context, class *Class) error

	// Subject operations
	ListSubjects(ctx context.Context, limit, offset int) ([]*Subject, int64, error)
	GetSubject(ctx context.Context, id string) (*Subject, error)
	CreateSubject(ctx context.Context, subject *Subject) error

	// Grade operations
	ListGradesByStudent(ctx context.Context, studentID string, limit, offset int) ([]*Grade, int64, error)
	GetGrade(ctx context.Context, id string) (*Grade, error)
	CreateGrade(ctx context.Context, grade *Grade) error
	UpdateGrade(ctx context.Context, grade *Grade) error

	// Attendance operations
	ListAttendanceByClass(ctx context.Context, classID string, startDate, endDate time.Time) ([]*Attendance, error)
	ListAttendanceByStudent(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*Attendance, error)
	MarkAttendance(ctx context.Context, attendance *Attendance) error
	BulkMarkAttendance(ctx context.Context, records []*Attendance) error

	// Assignment operations
	ListAssignments(ctx context.Context, classID string, limit, offset int) ([]*Assignment, int64, error)
	GetAssignment(ctx context.Context, id string) (*Assignment, error)
	CreateAssignment(ctx context.Context, assignment *Assignment) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ListAcademicYears retrieves a paginated list of academic years
func (r *PostgresRepository) ListAcademicYears(ctx context.Context, limit, offset int) ([]*AcademicYear, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM academic_years`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, school_id, name, start_date, end_date, is_current,
			created_at, updated_at
		FROM academic_years
		ORDER BY start_date DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var years []*AcademicYear
	for rows.Next() {
		year := &AcademicYear{}
		err := rows.Scan(
			&year.ID, &year.TenantID, &year.SchoolID, &year.Name,
			&year.StartDate, &year.EndDate, &year.IsCurrent,
			&year.CreatedAt, &year.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		years = append(years, year)
	}

	return years, total, nil
}

// GetAcademicYear retrieves an academic year by ID
func (r *PostgresRepository) GetAcademicYear(ctx context.Context, id string) (*AcademicYear, error) {
	query := `
		SELECT id, tenant_id, school_id, name, start_date, end_date, is_current,
			created_at, updated_at
		FROM academic_years
		WHERE id = $1
	`

	year := &AcademicYear{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&year.ID, &year.TenantID, &year.SchoolID, &year.Name,
		&year.StartDate, &year.EndDate, &year.IsCurrent,
		&year.CreatedAt, &year.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAcademicYearNotFound
		}
		return nil, err
	}

	return year, nil
}

// CreateAcademicYear creates a new academic year
func (r *PostgresRepository) CreateAcademicYear(ctx context.Context, year *AcademicYear) error {
	query := `
		INSERT INTO academic_years (
			id, tenant_id, school_id, name, start_date, end_date,
			is_current, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		year.ID, year.TenantID, year.SchoolID, year.Name,
		year.StartDate, year.EndDate, year.IsCurrent,
		time.Now(), time.Now(),
	)

	return err
}

// ListClasses retrieves a paginated list of classes
func (r *PostgresRepository) ListClasses(ctx context.Context, limit, offset int) ([]*Class, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM classes`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, school_id, grade_level_id, department_id,
			name, section, class_teacher_id, capacity, room_number,
			academic_year_id, created_at, updated_at
		FROM classes
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var classes []*Class
	for rows.Next() {
		class := &Class{}
		err := rows.Scan(
			&class.ID, &class.TenantID, &class.SchoolID, &class.GradeLevelID,
			&class.DepartmentID, &class.Name, &class.Section, &class.ClassTeacherID,
			&class.Capacity, &class.RoomNumber, &class.AcademicYearID,
			&class.CreatedAt, &class.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		classes = append(classes, class)
	}

	return classes, total, nil
}

// GetClass retrieves a class by ID
func (r *PostgresRepository) GetClass(ctx context.Context, id string) (*Class, error) {
	query := `
		SELECT id, tenant_id, school_id, grade_level_id, department_id,
			name, section, class_teacher_id, capacity, room_number,
			academic_year_id, created_at, updated_at
		FROM classes
		WHERE id = $1
	`

	class := &Class{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&class.ID, &class.TenantID, &class.SchoolID, &class.GradeLevelID,
		&class.DepartmentID, &class.Name, &class.Section, &class.ClassTeacherID,
		&class.Capacity, &class.RoomNumber, &class.AcademicYearID,
		&class.CreatedAt, &class.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrClassNotFound
		}
		return nil, err
	}

	return class, nil
}

// CreateClass creates a new class
func (r *PostgresRepository) CreateClass(ctx context.Context, class *Class) error {
	query := `
		INSERT INTO classes (
			id, tenant_id, school_id, grade_level_id, department_id,
			name, section, class_teacher_id, capacity, room_number,
			academic_year_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(ctx, query,
		class.ID, class.TenantID, class.SchoolID, class.GradeLevelID,
		class.DepartmentID, class.Name, class.Section, class.ClassTeacherID,
		class.Capacity, class.RoomNumber, class.AcademicYearID,
		time.Now(), time.Now(),
	)

	return err
}

// UpdateClass updates an existing class
func (r *PostgresRepository) UpdateClass(ctx context.Context, class *Class) error {
	query := `
		UPDATE classes SET
			name = $1, section = $2, class_teacher_id = $3,
			capacity = $4, room_number = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.Exec(ctx, query,
		class.Name, class.Section, class.ClassTeacherID,
		class.Capacity, class.RoomNumber, time.Now(), class.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrClassNotFound
	}

	return nil
}

// ListSubjects retrieves a paginated list of subjects
func (r *PostgresRepository) ListSubjects(ctx context.Context, limit, offset int) ([]*Subject, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM subjects`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, school_id, name, code, description,
			department_id, credit_hours, is_elective, created_at, updated_at
		FROM subjects
		ORDER BY name
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subjects []*Subject
	for rows.Next() {
		subject := &Subject{}
		err := rows.Scan(
			&subject.ID, &subject.TenantID, &subject.SchoolID, &subject.Name,
			&subject.Code, &subject.Description, &subject.DepartmentID,
			&subject.CreditHours, &subject.IsElective, &subject.CreatedAt,
			&subject.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		subjects = append(subjects, subject)
	}

	return subjects, total, nil
}

// GetSubject retrieves a subject by ID
func (r *PostgresRepository) GetSubject(ctx context.Context, id string) (*Subject, error) {
	query := `
		SELECT id, tenant_id, school_id, name, code, description,
			department_id, credit_hours, is_elective, created_at, updated_at
		FROM subjects
		WHERE id = $1
	`

	subject := &Subject{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&subject.ID, &subject.TenantID, &subject.SchoolID, &subject.Name,
		&subject.Code, &subject.Description, &subject.DepartmentID,
		&subject.CreditHours, &subject.IsElective, &subject.CreatedAt,
		&subject.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrSubjectNotFound
		}
		return nil, err
	}

	return subject, nil
}

// CreateSubject creates a new subject
func (r *PostgresRepository) CreateSubject(ctx context.Context, subject *Subject) error {
	query := `
		INSERT INTO subjects (
			id, tenant_id, school_id, name, code, description,
			department_id, credit_hours, is_elective, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		subject.ID, subject.TenantID, subject.SchoolID, subject.Name,
		subject.Code, subject.Description, subject.DepartmentID,
		subject.CreditHours, subject.IsElective, time.Now(), time.Now(),
	)

	return err
}

// ListGradesByStudent retrieves grades for a student
func (r *PostgresRepository) ListGradesByStudent(ctx context.Context, studentID string, limit, offset int) ([]*Grade, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM grades WHERE student_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, studentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, student_id, class_id, subject_id, term_id,
			assessment_type, assessment_name, score, max_score, grade, grade_point,
			remarks, graded_by, graded_at, created_at, updated_at
		FROM grades
		WHERE student_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, studentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var grades []*Grade
	for rows.Next() {
		grade := &Grade{}
		err := rows.Scan(
			&grade.ID, &grade.TenantID, &grade.StudentID, &grade.ClassID,
			&grade.SubjectID, &grade.TermID, &grade.AssessmentType, &grade.AssessmentName,
			&grade.Score, &grade.MaxScore, &grade.Grade, &grade.GradePoint,
			&grade.Remarks, &grade.GradedBy, &grade.GradedAt, &grade.CreatedAt,
			&grade.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		grades = append(grades, grade)
	}

	return grades, total, nil
}

// GetGrade retrieves a grade by ID
func (r *PostgresRepository) GetGrade(ctx context.Context, id string) (*Grade, error) {
	query := `
		SELECT id, tenant_id, student_id, class_id, subject_id, term_id,
			assessment_type, assessment_name, score, max_score, grade, grade_point,
			remarks, graded_by, graded_at, created_at, updated_at
		FROM grades
		WHERE id = $1
	`

	grade := &Grade{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&grade.ID, &grade.TenantID, &grade.StudentID, &grade.ClassID,
		&grade.SubjectID, &grade.TermID, &grade.AssessmentType, &grade.AssessmentName,
		&grade.Score, &grade.MaxScore, &grade.Grade, &grade.GradePoint,
		&grade.Remarks, &grade.GradedBy, &grade.GradedAt, &grade.CreatedAt,
		&grade.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrGradeNotFound
		}
		return nil, err
	}

	return grade, nil
}

// CreateGrade creates a new grade
func (r *PostgresRepository) CreateGrade(ctx context.Context, grade *Grade) error {
	query := `
		INSERT INTO grades (
			id, tenant_id, student_id, class_id, subject_id, term_id,
			assessment_type, assessment_name, score, max_score, grade, grade_point,
			remarks, graded_by, graded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.Exec(ctx, query,
		grade.ID, grade.TenantID, grade.StudentID, grade.ClassID,
		grade.SubjectID, grade.TermID, grade.AssessmentType, grade.AssessmentName,
		grade.Score, grade.MaxScore, grade.Grade, grade.GradePoint,
		grade.Remarks, grade.GradedBy, time.Now(), time.Now(), time.Now(),
	)

	return err
}

// UpdateGrade updates an existing grade
func (r *PostgresRepository) UpdateGrade(ctx context.Context, grade *Grade) error {
	query := `
		UPDATE grades SET
			score = $1, max_score = $2, grade = $3, grade_point = $4,
			remarks = $5, graded_by = $6, graded_at = $7, updated_at = $8
		WHERE id = $9
	`

	result, err := r.db.Exec(ctx, query,
		grade.Score, grade.MaxScore, grade.Grade, grade.GradePoint,
		grade.Remarks, grade.GradedBy, time.Now(), time.Now(), grade.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrGradeNotFound
	}

	return nil
}

// ListAttendanceByClass retrieves attendance for a class in date range
func (r *PostgresRepository) ListAttendanceByClass(ctx context.Context, classID string, startDate, endDate time.Time) ([]*Attendance, error) {
	query := `
		SELECT id, tenant_id, student_id, class_id, date, status,
			remarks, marked_by, marked_at, created_at, updated_at
		FROM attendance
		WHERE class_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date DESC, student_id
	`

	rows, err := r.db.Query(ctx, query, classID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*Attendance
	for rows.Next() {
		record := &Attendance{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.StudentID, &record.ClassID,
			&record.Date, &record.Status, &record.Remarks, &record.MarkedBy,
			&record.MarkedAt, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// ListAttendanceByStudent retrieves attendance for a student in date range
func (r *PostgresRepository) ListAttendanceByStudent(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*Attendance, error) {
	query := `
		SELECT id, tenant_id, student_id, class_id, date, status,
			remarks, marked_by, marked_at, created_at, updated_at
		FROM attendance
		WHERE student_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date DESC
	`

	rows, err := r.db.Query(ctx, query, studentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*Attendance
	for rows.Next() {
		record := &Attendance{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.StudentID, &record.ClassID,
			&record.Date, &record.Status, &record.Remarks, &record.MarkedBy,
			&record.MarkedAt, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// MarkAttendance creates or updates an attendance record
func (r *PostgresRepository) MarkAttendance(ctx context.Context, attendance *Attendance) error {
	query := `
		INSERT INTO attendance (
			id, tenant_id, student_id, class_id, date, status,
			remarks, marked_by, marked_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, student_id, class_id, date)
		DO UPDATE SET
			status = EXCLUDED.status,
			remarks = EXCLUDED.remarks,
			marked_by = EXCLUDED.marked_by,
			marked_at = EXCLUDED.marked_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		attendance.ID, attendance.TenantID, attendance.StudentID, attendance.ClassID,
		attendance.Date, attendance.Status, attendance.Remarks, attendance.MarkedBy,
		time.Now(), time.Now(), time.Now(),
	)

	return err
}

// BulkMarkAttendance creates or updates multiple attendance records
func (r *PostgresRepository) BulkMarkAttendance(ctx context.Context, records []*Attendance) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO attendance (
			id, tenant_id, student_id, class_id, date, status,
			remarks, marked_by, marked_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, student_id, class_id, date)
		DO UPDATE SET
			status = EXCLUDED.status,
			remarks = EXCLUDED.remarks,
			marked_by = EXCLUDED.marked_by,
			marked_at = EXCLUDED.marked_at,
			updated_at = EXCLUDED.updated_at
	`

	for _, record := range records {
		_, err := tx.Exec(ctx, query,
			record.ID, record.TenantID, record.StudentID, record.ClassID,
			record.Date, record.Status, record.Remarks, record.MarkedBy,
			time.Now(), time.Now(), time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// ListAssignments retrieves assignments for a class
func (r *PostgresRepository) ListAssignments(ctx context.Context, classID string, limit, offset int) ([]*Assignment, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM assignments WHERE class_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, classID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, class_id, subject_id, teacher_id, title,
			description, assignment_type, max_score, due_date, assigned_date,
			is_published, created_at, updated_at
		FROM assignments
		WHERE class_id = $1
		ORDER BY due_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, classID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var assignments []*Assignment
	for rows.Next() {
		assignment := &Assignment{}
		err := rows.Scan(
			&assignment.ID, &assignment.TenantID, &assignment.ClassID,
			&assignment.SubjectID, &assignment.TeacherID, &assignment.Title,
			&assignment.Description, &assignment.AssignmentType, &assignment.MaxScore,
			&assignment.DueDate, &assignment.AssignedDate, &assignment.IsPublished,
			&assignment.CreatedAt, &assignment.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		assignments = append(assignments, assignment)
	}

	return assignments, total, nil
}

// GetAssignment retrieves an assignment by ID
func (r *PostgresRepository) GetAssignment(ctx context.Context, id string) (*Assignment, error) {
	query := `
		SELECT id, tenant_id, class_id, subject_id, teacher_id, title,
			description, assignment_type, max_score, due_date, assigned_date,
			is_published, created_at, updated_at
		FROM assignments
		WHERE id = $1
	`

	assignment := &Assignment{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&assignment.ID, &assignment.TenantID, &assignment.ClassID,
		&assignment.SubjectID, &assignment.TeacherID, &assignment.Title,
		&assignment.Description, &assignment.AssignmentType, &assignment.MaxScore,
		&assignment.DueDate, &assignment.AssignedDate, &assignment.IsPublished,
		&assignment.CreatedAt, &assignment.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAssignmentNotFound
		}
		return nil, err
	}

	return assignment, nil
}

// CreateAssignment creates a new assignment
func (r *PostgresRepository) CreateAssignment(ctx context.Context, assignment *Assignment) error {
	query := `
		INSERT INTO assignments (
			id, tenant_id, class_id, subject_id, teacher_id, title,
			description, assignment_type, max_score, due_date, assigned_date,
			is_published, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Exec(ctx, query,
		assignment.ID, assignment.TenantID, assignment.ClassID,
		assignment.SubjectID, assignment.TeacherID, assignment.Title,
		assignment.Description, assignment.AssignmentType, assignment.MaxScore,
		assignment.DueDate, time.Now(), assignment.IsPublished,
		time.Now(), time.Now(),
	)

	return err
}
