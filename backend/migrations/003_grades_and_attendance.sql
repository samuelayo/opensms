-- +goose Up
-- Grades and Attendance Schema

-- ============================================================================
-- GRADING SYSTEM
-- ============================================================================

-- Grading Scales
CREATE TABLE grading_scales (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    school_id UUID REFERENCES schools(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, school_id, name)
);

CREATE INDEX idx_grading_scales_tenant_id ON grading_scales(tenant_id);
CREATE INDEX idx_grading_scales_school_id ON grading_scales(school_id);

ALTER TABLE grading_scales ENABLE ROW LEVEL SECURITY;
CREATE POLICY grading_scales_tenant_isolation ON grading_scales
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Grade Entries (A, B, C, etc.)
CREATE TABLE grade_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    grading_scale_id UUID NOT NULL REFERENCES grading_scales(id) ON DELETE CASCADE,
    grade VARCHAR(10) NOT NULL, -- A+, A, B+, B, etc.
    min_score DECIMAL(5,2) NOT NULL,
    max_score DECIMAL(5,2) NOT NULL,
    grade_point DECIMAL(3,2), -- GPA value
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_grade_entries_grading_scale_id ON grade_entries(grading_scale_id);

-- Student Grades
CREATE TABLE grades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    assessment_type VARCHAR(50) NOT NULL, -- exam, quiz, assignment, project, midterm, final
    assessment_name VARCHAR(255),
    score DECIMAL(5,2) NOT NULL,
    max_score DECIMAL(5,2) NOT NULL DEFAULT 100,
    grade VARCHAR(10), -- A, B+, etc.
    grade_point DECIMAL(3,2),
    remarks TEXT,
    graded_by UUID REFERENCES users(id),
    graded_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_grades_tenant_id ON grades(tenant_id);
CREATE INDEX idx_grades_student_id ON grades(student_id);
CREATE INDEX idx_grades_class_id ON grades(class_id);
CREATE INDEX idx_grades_subject_id ON grades(subject_id);
CREATE INDEX idx_grades_term_id ON grades(term_id);

ALTER TABLE grades ENABLE ROW LEVEL SECURITY;
CREATE POLICY grades_tenant_isolation ON grades
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- ============================================================================
-- ATTENDANCE SYSTEM
-- ============================================================================

-- Attendance Records
CREATE TABLE attendance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    status attendance_status NOT NULL DEFAULT 'present',
    remarks TEXT,
    marked_by UUID REFERENCES users(id),
    marked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, student_id, class_id, date)
);

CREATE INDEX idx_attendance_tenant_id ON attendance(tenant_id);
CREATE INDEX idx_attendance_student_id ON attendance(student_id);
CREATE INDEX idx_attendance_class_id ON attendance(class_id);
CREATE INDEX idx_attendance_date ON attendance(date);

ALTER TABLE attendance ENABLE ROW LEVEL SECURITY;
CREATE POLICY attendance_tenant_isolation ON attendance
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Attendance Summary (Materialized for performance)
CREATE MATERIALIZED VIEW attendance_summary AS
SELECT
    tenant_id,
    student_id,
    class_id,
    DATE_TRUNC('month', date) as month,
    COUNT(*) FILTER (WHERE status = 'present') as days_present,
    COUNT(*) FILTER (WHERE status = 'absent') as days_absent,
    COUNT(*) FILTER (WHERE status = 'late') as days_late,
    COUNT(*) FILTER (WHERE status = 'excused') as days_excused,
    COUNT(*) as total_days,
    ROUND((COUNT(*) FILTER (WHERE status = 'present')::DECIMAL / NULLIF(COUNT(*), 0)) * 100, 2) as attendance_percentage
FROM attendance
GROUP BY tenant_id, student_id, class_id, DATE_TRUNC('month', date);

CREATE INDEX idx_attendance_summary_student ON attendance_summary(student_id);
CREATE INDEX idx_attendance_summary_class ON attendance_summary(class_id);

-- ============================================================================
-- ASSIGNMENTS AND ASSESSMENTS
-- ============================================================================

-- Assignments
CREATE TABLE assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assignment_type VARCHAR(50) NOT NULL, -- homework, project, essay, lab
    max_score DECIMAL(5,2) NOT NULL DEFAULT 100,
    due_date TIMESTAMP NOT NULL,
    assigned_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_published BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_assignments_tenant_id ON assignments(tenant_id);
CREATE INDEX idx_assignments_class_id ON assignments(class_id);
CREATE INDEX idx_assignments_subject_id ON assignments(subject_id);
CREATE INDEX idx_assignments_teacher_id ON assignments(teacher_id);
CREATE INDEX idx_assignments_due_date ON assignments(due_date);

ALTER TABLE assignments ENABLE ROW LEVEL SECURITY;
CREATE POLICY assignments_tenant_isolation ON assignments
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Assignment Submissions
CREATE TABLE assignment_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assignment_id UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    submission_text TEXT,
    file_path TEXT,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    score DECIMAL(5,2),
    feedback TEXT,
    graded_at TIMESTAMP,
    graded_by UUID REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'submitted', -- submitted, graded, late, missing
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(assignment_id, student_id)
);

CREATE INDEX idx_assignment_submissions_assignment_id ON assignment_submissions(assignment_id);
CREATE INDEX idx_assignment_submissions_student_id ON assignment_submissions(student_id);

-- ============================================================================
-- EXAMS AND ASSESSMENTS
-- ============================================================================

-- Exams
CREATE TABLE exams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    exam_name VARCHAR(255) NOT NULL,
    exam_type VARCHAR(50) NOT NULL, -- midterm, final, quiz, unit_test
    exam_date DATE NOT NULL,
    start_time TIME,
    duration_minutes INT,
    max_score DECIMAL(5,2) NOT NULL DEFAULT 100,
    passing_score DECIMAL(5,2),
    room_number VARCHAR(50),
    instructions TEXT,
    is_published BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exams_tenant_id ON exams(tenant_id);
CREATE INDEX idx_exams_class_id ON exams(class_id);
CREATE INDEX idx_exams_subject_id ON exams(subject_id);
CREATE INDEX idx_exams_term_id ON exams(term_id);
CREATE INDEX idx_exams_exam_date ON exams(exam_date);

ALTER TABLE exams ENABLE ROW LEVEL SECURITY;
CREATE POLICY exams_tenant_isolation ON exams
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Exam Results
CREATE TABLE exam_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    score DECIMAL(5,2) NOT NULL,
    grade VARCHAR(10),
    grade_point DECIMAL(3,2),
    remarks TEXT,
    is_absent BOOLEAN DEFAULT false,
    graded_by UUID REFERENCES users(id),
    graded_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(exam_id, student_id)
);

CREATE INDEX idx_exam_results_exam_id ON exam_results(exam_id);
CREATE INDEX idx_exam_results_student_id ON exam_results(student_id);

-- ============================================================================
-- TIMETABLE/SCHEDULE
-- ============================================================================

-- Class Schedules
CREATE TABLE class_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_id UUID REFERENCES teachers(id),
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL, -- 1=Monday, 7=Sunday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    room_number VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_class_schedules_tenant_id ON class_schedules(tenant_id);
CREATE INDEX idx_class_schedules_class_id ON class_schedules(class_id);
CREATE INDEX idx_class_schedules_teacher_id ON class_schedules(teacher_id);
CREATE INDEX idx_class_schedules_day ON class_schedules(day_of_week);

ALTER TABLE class_schedules ENABLE ROW LEVEL SECURITY;
CREATE POLICY class_schedules_tenant_isolation ON class_schedules
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- +goose Down
DROP MATERIALIZED VIEW IF EXISTS attendance_summary;
DROP TABLE IF EXISTS class_schedules CASCADE;
DROP TABLE IF EXISTS exam_results CASCADE;
DROP TABLE IF EXISTS exams CASCADE;
DROP TABLE IF EXISTS assignment_submissions CASCADE;
DROP TABLE IF EXISTS assignments CASCADE;
DROP TABLE IF EXISTS attendance CASCADE;
DROP TABLE IF EXISTS grades CASCADE;
DROP TABLE IF EXISTS grade_entries CASCADE;
DROP TABLE IF EXISTS grading_scales CASCADE;
