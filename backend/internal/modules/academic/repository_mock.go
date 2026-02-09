package academic

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	mock.Mock
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// Academic Year operations
func (m *MockRepository) ListAcademicYears(ctx context.Context, limit, offset int) ([]*AcademicYear, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*AcademicYear), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetAcademicYearByID(ctx context.Context, id string) (*AcademicYear, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AcademicYear), args.Error(1)
}

func (m *MockRepository) CreateAcademicYear(ctx context.Context, year *AcademicYear) error {
	args := m.Called(ctx, year)
	return args.Error(0)
}

func (m *MockRepository) UpdateAcademicYear(ctx context.Context, year *AcademicYear) error {
	args := m.Called(ctx, year)
	return args.Error(0)
}

// Class operations
func (m *MockRepository) ListClasses(ctx context.Context, limit, offset int) ([]*Class, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Class), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetClassByID(ctx context.Context, id string) (*Class, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Class), args.Error(1)
}

func (m *MockRepository) CreateClass(ctx context.Context, class *Class) error {
	args := m.Called(ctx, class)
	return args.Error(0)
}

func (m *MockRepository) UpdateClass(ctx context.Context, class *Class) error {
	args := m.Called(ctx, class)
	return args.Error(0)
}

// Subject operations
func (m *MockRepository) ListSubjects(ctx context.Context, limit, offset int) ([]*Subject, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Subject), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetSubjectByID(ctx context.Context, id string) (*Subject, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Subject), args.Error(1)
}

func (m *MockRepository) CreateSubject(ctx context.Context, subject *Subject) error {
	args := m.Called(ctx, subject)
	return args.Error(0)
}

func (m *MockRepository) UpdateSubject(ctx context.Context, subject *Subject) error {
	args := m.Called(ctx, subject)
	return args.Error(0)
}

// Grade operations
func (m *MockRepository) ListGradesByStudent(ctx context.Context, studentID string, limit, offset int) ([]*Grade, int64, error) {
	args := m.Called(ctx, studentID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Grade), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetGradeByID(ctx context.Context, id string) (*Grade, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Grade), args.Error(1)
}

func (m *MockRepository) CreateGrade(ctx context.Context, grade *Grade) error {
	args := m.Called(ctx, grade)
	return args.Error(0)
}

func (m *MockRepository) UpdateGrade(ctx context.Context, grade *Grade) error {
	args := m.Called(ctx, grade)
	return args.Error(0)
}

// Attendance operations
func (m *MockRepository) ListAttendanceByClass(ctx context.Context, classID string, startDate, endDate time.Time) ([]*Attendance, error) {
	args := m.Called(ctx, classID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Attendance), args.Error(1)
}

func (m *MockRepository) ListAttendanceByStudent(ctx context.Context, studentID, classID string, startDate, endDate time.Time) ([]*Attendance, error) {
	args := m.Called(ctx, studentID, classID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Attendance), args.Error(1)
}

func (m *MockRepository) MarkAttendance(ctx context.Context, attendance *Attendance) error {
	args := m.Called(ctx, attendance)
	return args.Error(0)
}

func (m *MockRepository) BulkMarkAttendance(ctx context.Context, records []*Attendance) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

// Assignment operations
func (m *MockRepository) ListAssignments(ctx context.Context, classID string, limit, offset int) ([]*Assignment, int64, error) {
	args := m.Called(ctx, classID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Assignment), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetAssignmentByID(ctx context.Context, id string) (*Assignment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Assignment), args.Error(1)
}

func (m *MockRepository) CreateAssignment(ctx context.Context, assignment *Assignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockRepository) UpdateAssignment(ctx context.Context, assignment *Assignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}
