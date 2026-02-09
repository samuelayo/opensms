package academic

import "time"

// CreateAcademicYearRequest represents a request to create an academic year
type CreateAcademicYearRequest struct {
	TenantID  string `json:"tenant_id" validate:"required,uuid"`
	Name      string `json:"name" validate:"required"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
	IsCurrent bool   `json:"is_current"`
}

// UpdateAcademicYearRequest represents a request to update an academic year
type UpdateAcademicYearRequest struct {
	Name      string `json:"name,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	IsCurrent *bool  `json:"is_current,omitempty"`
}

// CreateGradeLevelRequest represents a request to create a grade level
type CreateGradeLevelRequest struct {
	TenantID    string `json:"tenant_id" validate:"required,uuid"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Level       int    `json:"level" validate:"required,min=1"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

// UpdateGradeLevelRequest represents a request to update a grade level
type UpdateGradeLevelRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Level       int    `json:"level,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

// CreateClassRequest represents a request to create a class
type CreateClassRequest struct {
	TenantID        string `json:"tenant_id" validate:"required,uuid"`
	SchoolID        string `json:"school_id" validate:"required,uuid"`
	GradeLevelID    string `json:"grade_level_id" validate:"required,uuid"`
	AcademicYearID  string `json:"academic_year_id" validate:"required,uuid"`
	Name            string `json:"name" validate:"required"`
	Section         string `json:"section,omitempty"`
	ClassTeacherID  string `json:"class_teacher_id,omitempty"`
	MaxStudents     int    `json:"max_students,omitempty"`
	ClassroomNumber string `json:"classroom_number,omitempty"`
}

// UpdateClassRequest represents a request to update a class
type UpdateClassRequest struct {
	Name            string `json:"name,omitempty"`
	Section         string `json:"section,omitempty"`
	ClassTeacherID  string `json:"class_teacher_id,omitempty"`
	MaxStudents     int    `json:"max_students,omitempty"`
	ClassroomNumber string `json:"classroom_number,omitempty"`
}

// CreateSubjectRequest represents a request to create a subject
type CreateSubjectRequest struct {
	TenantID    string `json:"tenant_id" validate:"required,uuid"`
	Code        string `json:"code" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Credits     int    `json:"credits,omitempty"`
	IsElective  bool   `json:"is_elective"`
}

// UpdateSubjectRequest represents a request to update a subject
type UpdateSubjectRequest struct {
	Code        string `json:"code,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Credits     int    `json:"credits,omitempty"`
	IsElective  *bool  `json:"is_elective,omitempty"`
}

// CreateGradeRequest represents a request to create a grade
type CreateGradeRequest struct {
	TenantID       string  `json:"tenant_id" validate:"required,uuid"`
	StudentID      string  `json:"student_id" validate:"required,uuid"`
	ClassID        string  `json:"class_id" validate:"required,uuid"`
	SubjectID      string  `json:"subject_id" validate:"required,uuid"`
	AssessmentType string  `json:"assessment_type" validate:"required,oneof=exam quiz assignment project midterm final"`
	Score          float64 `json:"score" validate:"required,min=0"`
	MaxScore       float64 `json:"max_score" validate:"required,min=0"`
	Grade          string  `json:"grade,omitempty"`
	GradePoint     float64 `json:"grade_point,omitempty"`
	Remarks        string  `json:"remarks,omitempty"`
	GradedDate     string  `json:"graded_date,omitempty"`
}

// UpdateGradeRequest represents a request to update a grade
type UpdateGradeRequest struct {
	Score      float64 `json:"score,omitempty"`
	MaxScore   float64 `json:"max_score,omitempty"`
	Grade      string  `json:"grade,omitempty"`
	GradePoint float64 `json:"grade_point,omitempty"`
	Remarks    string  `json:"remarks,omitempty"`
	GradedDate string  `json:"graded_date,omitempty"`
}

// MarkAttendanceRequest represents a request to mark attendance
type MarkAttendanceRequest struct {
	TenantID  string `json:"tenant_id" validate:"required,uuid"`
	StudentID string `json:"student_id" validate:"required,uuid"`
	ClassID   string `json:"class_id" validate:"required,uuid"`
	Date      string `json:"date" validate:"required"`
	Status    string `json:"status" validate:"required,oneof=present absent late excused"`
	Remarks   string `json:"remarks,omitempty"`
}

// BulkAttendanceRequest represents a request to mark attendance for multiple students
type BulkAttendanceRequest struct {
	TenantID string                  `json:"tenant_id" validate:"required,uuid"`
	ClassID  string                  `json:"class_id" validate:"required,uuid"`
	Date     string                  `json:"date" validate:"required"`
	Records  []BulkAttendanceRecord  `json:"records" validate:"required,min=1,dive"`
}

// BulkAttendanceRecord represents a single student's attendance in bulk marking
type BulkAttendanceRecord struct {
	StudentID string `json:"student_id" validate:"required,uuid"`
	Status    string `json:"status" validate:"required,oneof=present absent late excused"`
	Remarks   string `json:"remarks,omitempty"`
}

// CreateAssignmentRequest represents a request to create an assignment
type CreateAssignmentRequest struct {
	TenantID       string  `json:"tenant_id" validate:"required,uuid"`
	ClassID        string  `json:"class_id" validate:"required,uuid"`
	SubjectID      string  `json:"subject_id" validate:"required,uuid"`
	TeacherID      string  `json:"teacher_id" validate:"required,uuid"`
	Title          string  `json:"title" validate:"required"`
	Description    string  `json:"description,omitempty"`
	AssignmentType string  `json:"assignment_type,omitempty"`
	MaxScore       float64 `json:"max_score" validate:"required,min=0"`
	DueDate        string  `json:"due_date" validate:"required"`
	AttachmentURL  string  `json:"attachment_url,omitempty"`
}

// UpdateAssignmentRequest represents a request to update an assignment
type UpdateAssignmentRequest struct {
	Title          string  `json:"title,omitempty"`
	Description    string  `json:"description,omitempty"`
	AssignmentType string  `json:"assignment_type,omitempty"`
	MaxScore       float64 `json:"max_score,omitempty"`
	DueDate        string  `json:"due_date,omitempty"`
	AttachmentURL  string  `json:"attachment_url,omitempty"`
}

// CreateExamRequest represents a request to create an exam
type CreateExamRequest struct {
	TenantID       string  `json:"tenant_id" validate:"required,uuid"`
	ClassID        string  `json:"class_id" validate:"required,uuid"`
	SubjectID      string  `json:"subject_id" validate:"required,uuid"`
	AcademicYearID string  `json:"academic_year_id" validate:"required,uuid"`
	Title          string  `json:"title" validate:"required"`
	ExamType       string  `json:"exam_type" validate:"required,oneof=midterm final quiz unit_test"`
	ExamDate       string  `json:"exam_date" validate:"required"`
	Duration       int     `json:"duration,omitempty"`
	MaxScore       float64 `json:"max_score" validate:"required,min=0"`
	PassingScore   float64 `json:"passing_score,omitempty"`
	Instructions   string  `json:"instructions,omitempty"`
}

// UpdateExamRequest represents a request to update an exam
type UpdateExamRequest struct {
	Title        string  `json:"title,omitempty"`
	ExamType     string  `json:"exam_type,omitempty"`
	ExamDate     string  `json:"exam_date,omitempty"`
	Duration     int     `json:"duration,omitempty"`
	MaxScore     float64 `json:"max_score,omitempty"`
	PassingScore float64 `json:"passing_score,omitempty"`
	Instructions string  `json:"instructions,omitempty"`
}

// CreateScheduleRequest represents a request to create a class schedule
type CreateScheduleRequest struct {
	TenantID    string `json:"tenant_id" validate:"required,uuid"`
	ClassID     string `json:"class_id" validate:"required,uuid"`
	SubjectID   string `json:"subject_id" validate:"required,uuid"`
	TeacherID   string `json:"teacher_id" validate:"required,uuid"`
	DayOfWeek   int    `json:"day_of_week" validate:"required,min=1,max=7"`
	StartTime   string `json:"start_time" validate:"required"`
	EndTime     string `json:"end_time" validate:"required"`
	RoomNumber  string `json:"room_number,omitempty"`
}

// UpdateScheduleRequest represents a request to update a class schedule
type UpdateScheduleRequest struct {
	SubjectID  string `json:"subject_id,omitempty"`
	TeacherID  string `json:"teacher_id,omitempty"`
	DayOfWeek  int    `json:"day_of_week,omitempty"`
	StartTime  string `json:"start_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
	RoomNumber string `json:"room_number,omitempty"`
}

// AcademicYearDTO represents an academic year in API responses
type AcademicYearDTO struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsCurrent bool   `json:"is_current"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GradeLevelDTO represents a grade level in API responses
type GradeLevelDTO struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Level       int    `json:"level"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ClassDTO represents a class in API responses
type ClassDTO struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	SchoolID        string `json:"school_id"`
	GradeLevelID    string `json:"grade_level_id"`
	AcademicYearID  string `json:"academic_year_id"`
	Name            string `json:"name"`
	Section         string `json:"section,omitempty"`
	ClassTeacherID  string `json:"class_teacher_id,omitempty"`
	MaxStudents     int    `json:"max_students"`
	CurrentStudents int    `json:"current_students"`
	ClassroomNumber string `json:"classroom_number,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// SubjectDTO represents a subject in API responses
type SubjectDTO struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Credits     int    `json:"credits"`
	IsElective  bool   `json:"is_elective"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GradeDTO represents a grade in API responses
type GradeDTO struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenant_id"`
	StudentID      string  `json:"student_id"`
	ClassID        string  `json:"class_id"`
	SubjectID      string  `json:"subject_id"`
	AssessmentType string  `json:"assessment_type"`
	Score          float64 `json:"score"`
	MaxScore       float64 `json:"max_score"`
	Percentage     float64 `json:"percentage"`
	Grade          string  `json:"grade,omitempty"`
	GradePoint     float64 `json:"grade_point,omitempty"`
	Remarks        string  `json:"remarks,omitempty"`
	GradedBy       string  `json:"graded_by,omitempty"`
	GradedDate     string  `json:"graded_date,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// AttendanceDTO represents attendance in API responses
type AttendanceDTO struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	StudentID string `json:"student_id"`
	ClassID   string `json:"class_id"`
	Date      string `json:"date"`
	Status    string `json:"status"`
	Remarks   string `json:"remarks,omitempty"`
	MarkedBy  string `json:"marked_by,omitempty"`
	MarkedAt  string `json:"marked_at,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AssignmentDTO represents an assignment in API responses
type AssignmentDTO struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenant_id"`
	ClassID        string  `json:"class_id"`
	SubjectID      string  `json:"subject_id"`
	TeacherID      string  `json:"teacher_id"`
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	AssignmentType string  `json:"assignment_type,omitempty"`
	MaxScore       float64 `json:"max_score"`
	DueDate        string  `json:"due_date"`
	AttachmentURL  string  `json:"attachment_url,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ExamDTO represents an exam in API responses
type ExamDTO struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenant_id"`
	ClassID        string  `json:"class_id"`
	SubjectID      string  `json:"subject_id"`
	AcademicYearID string  `json:"academic_year_id"`
	Title          string  `json:"title"`
	ExamType       string  `json:"exam_type"`
	ExamDate       string  `json:"exam_date"`
	Duration       int     `json:"duration"`
	MaxScore       float64 `json:"max_score"`
	PassingScore   float64 `json:"passing_score"`
	Instructions   string  `json:"instructions,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ClassScheduleDTO represents a class schedule in API responses
type ClassScheduleDTO struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	ClassID    string `json:"class_id"`
	SubjectID  string `json:"subject_id"`
	TeacherID  string `json:"teacher_id"`
	DayOfWeek  int    `json:"day_of_week"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	RoomNumber string `json:"room_number,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// AttendanceStatsDTO represents attendance statistics
type AttendanceStatsDTO struct {
	StudentID      string  `json:"student_id"`
	ClassID        string  `json:"class_id"`
	TotalDays      int     `json:"total_days"`
	PresentDays    int     `json:"present_days"`
	AbsentDays     int     `json:"absent_days"`
	LateDays       int     `json:"late_days"`
	ExcusedDays    int     `json:"excused_days"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// GradeStatsDTO represents grade statistics for a student
type GradeStatsDTO struct {
	StudentID    string  `json:"student_id"`
	SubjectID    string  `json:"subject_id"`
	AverageScore float64 `json:"average_score"`
	HighestScore float64 `json:"highest_score"`
	LowestScore  float64 `json:"lowest_score"`
	TotalGrades  int     `json:"total_grades"`
	LetterGrade  string  `json:"letter_grade,omitempty"`
	GPA          float64 `json:"gpa,omitempty"`
}

// ToAcademicYearDTO converts an AcademicYear to AcademicYearDTO
func ToAcademicYearDTO(year *AcademicYear) *AcademicYearDTO {
	return &AcademicYearDTO{
		ID:        year.ID,
		TenantID:  year.TenantID,
		Name:      year.Name,
		StartDate: year.StartDate.Format("2006-01-02"),
		EndDate:   year.EndDate.Format("2006-01-02"),
		IsCurrent: year.IsCurrent,
		CreatedAt: year.CreatedAt.Format(time.RFC3339),
		UpdatedAt: year.UpdatedAt.Format(time.RFC3339),
	}
}

// ToGradeLevelDTO converts a GradeLevel to GradeLevelDTO
func ToGradeLevelDTO(level *GradeLevel) *GradeLevelDTO {
	dto := &GradeLevelDTO{
		ID:        level.ID,
		TenantID:  level.TenantID,
		Name:      level.Name,
		Level:     level.Level,
		SortOrder: level.SortOrder,
		CreatedAt: level.CreatedAt.Format(time.RFC3339),
		UpdatedAt: level.UpdatedAt.Format(time.RFC3339),
	}
	if level.Description != nil {
		dto.Description = *level.Description
	}
	return dto
}

// ToClassDTO converts a Class to ClassDTO
func ToClassDTO(class *Class) *ClassDTO {
	dto := &ClassDTO{
		ID:              class.ID,
		TenantID:        class.TenantID,
		SchoolID:        class.SchoolID,
		GradeLevelID:    class.GradeLevelID,
		AcademicYearID:  class.AcademicYearID,
		Name:            class.Name,
		MaxStudents:     class.MaxStudents,
		CurrentStudents: class.CurrentStudents,
		CreatedAt:       class.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       class.UpdatedAt.Format(time.RFC3339),
	}
	if class.Section != nil {
		dto.Section = *class.Section
	}
	if class.ClassTeacherID != nil {
		dto.ClassTeacherID = *class.ClassTeacherID
	}
	if class.ClassroomNumber != nil {
		dto.ClassroomNumber = *class.ClassroomNumber
	}
	return dto
}

// ToSubjectDTO converts a Subject to SubjectDTO
func ToSubjectDTO(subject *Subject) *SubjectDTO {
	dto := &SubjectDTO{
		ID:         subject.ID,
		TenantID:   subject.TenantID,
		Code:       subject.Code,
		Name:       subject.Name,
		Credits:    subject.Credits,
		IsElective: subject.IsElective,
		CreatedAt:  subject.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  subject.UpdatedAt.Format(time.RFC3339),
	}
	if subject.Description != nil {
		dto.Description = *subject.Description
	}
	return dto
}

// ToGradeDTO converts a Grade to GradeDTO
func ToGradeDTO(grade *Grade) *GradeDTO {
	dto := &GradeDTO{
		ID:             grade.ID,
		TenantID:       grade.TenantID,
		StudentID:      grade.StudentID,
		ClassID:        grade.ClassID,
		SubjectID:      grade.SubjectID,
		AssessmentType: grade.AssessmentType,
		Score:          grade.Score,
		MaxScore:       grade.MaxScore,
		Percentage:     (grade.Score / grade.MaxScore) * 100,
		CreatedAt:      grade.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      grade.UpdatedAt.Format(time.RFC3339),
	}
	if grade.Grade != nil {
		dto.Grade = *grade.Grade
	}
	if grade.GradePoint != nil {
		dto.GradePoint = *grade.GradePoint
	}
	if grade.Remarks != nil {
		dto.Remarks = *grade.Remarks
	}
	if grade.GradedBy != nil {
		dto.GradedBy = *grade.GradedBy
	}
	if grade.GradedDate != nil {
		dto.GradedDate = grade.GradedDate.Format("2006-01-02")
	}
	return dto
}

// ToAttendanceDTO converts an Attendance to AttendanceDTO
func ToAttendanceDTO(attendance *Attendance) *AttendanceDTO {
	dto := &AttendanceDTO{
		ID:        attendance.ID,
		TenantID:  attendance.TenantID,
		StudentID: attendance.StudentID,
		ClassID:   attendance.ClassID,
		Date:      attendance.Date.Format("2006-01-02"),
		Status:    attendance.Status,
		CreatedAt: attendance.CreatedAt.Format(time.RFC3339),
		UpdatedAt: attendance.UpdatedAt.Format(time.RFC3339),
	}
	if attendance.Remarks != nil {
		dto.Remarks = *attendance.Remarks
	}
	if attendance.MarkedBy != nil {
		dto.MarkedBy = *attendance.MarkedBy
	}
	if attendance.MarkedAt != nil {
		dto.MarkedAt = attendance.MarkedAt.Format(time.RFC3339)
	}
	return dto
}

// ToAssignmentDTO converts an Assignment to AssignmentDTO
func ToAssignmentDTO(assignment *Assignment) *AssignmentDTO {
	dto := &AssignmentDTO{
		ID:        assignment.ID,
		TenantID:  assignment.TenantID,
		ClassID:   assignment.ClassID,
		SubjectID: assignment.SubjectID,
		TeacherID: assignment.TeacherID,
		Title:     assignment.Title,
		MaxScore:  assignment.MaxScore,
		DueDate:   assignment.DueDate.Format(time.RFC3339),
		CreatedAt: assignment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: assignment.UpdatedAt.Format(time.RFC3339),
	}
	if assignment.Description != nil {
		dto.Description = *assignment.Description
	}
	if assignment.AssignmentType != nil {
		dto.AssignmentType = *assignment.AssignmentType
	}
	if assignment.AttachmentURL != nil {
		dto.AttachmentURL = *assignment.AttachmentURL
	}
	return dto
}

// ToExamDTO converts an Exam to ExamDTO
func ToExamDTO(exam *Exam) *ExamDTO {
	dto := &ExamDTO{
		ID:             exam.ID,
		TenantID:       exam.TenantID,
		ClassID:        exam.ClassID,
		SubjectID:      exam.SubjectID,
		AcademicYearID: exam.AcademicYearID,
		Title:          exam.Title,
		ExamType:       exam.ExamType,
		ExamDate:       exam.ExamDate.Format("2006-01-02"),
		Duration:       exam.Duration,
		MaxScore:       exam.MaxScore,
		PassingScore:   exam.PassingScore,
		CreatedAt:      exam.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      exam.UpdatedAt.Format(time.RFC3339),
	}
	if exam.Instructions != nil {
		dto.Instructions = *exam.Instructions
	}
	return dto
}

// ToClassScheduleDTO converts a ClassSchedule to ClassScheduleDTO
func ToClassScheduleDTO(schedule *ClassSchedule) *ClassScheduleDTO {
	dto := &ClassScheduleDTO{
		ID:        schedule.ID,
		TenantID:  schedule.TenantID,
		ClassID:   schedule.ClassID,
		SubjectID: schedule.SubjectID,
		TeacherID: schedule.TeacherID,
		DayOfWeek: schedule.DayOfWeek,
		StartTime: schedule.StartTime,
		EndTime:   schedule.EndTime,
		CreatedAt: schedule.CreatedAt.Format(time.RFC3339),
		UpdatedAt: schedule.UpdatedAt.Format(time.RFC3339),
	}
	if schedule.RoomNumber != nil {
		dto.RoomNumber = *schedule.RoomNumber
	}
	return dto
}
