-- +goose Up
-- Students, Teachers, and Enrollment Schema

-- ============================================================================
-- STUDENT MANAGEMENT
-- ============================================================================

-- Students
CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    admission_number VARCHAR(50) UNIQUE NOT NULL,
    admission_date DATE NOT NULL,
    date_of_birth DATE NOT NULL,
    gender gender NOT NULL,
    blood_group VARCHAR(10),
    nationality VARCHAR(2), -- ISO 3166-1 alpha-2
    religion VARCHAR(100),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    emergency_contact_name VARCHAR(255),
    emergency_contact_phone VARCHAR(50),
    emergency_contact_relationship VARCHAR(100),
    medical_conditions TEXT,
    allergies TEXT,
    previous_school VARCHAR(255),
    current_grade_level_id UUID REFERENCES grade_levels(id),
    current_class_id UUID REFERENCES classes(id),
    is_active BOOLEAN DEFAULT true,
    withdrawal_date DATE,
    withdrawal_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, admission_number)
);

CREATE INDEX idx_students_tenant_id ON students(tenant_id);
CREATE INDEX idx_students_user_id ON students(user_id);
CREATE INDEX idx_students_school_id ON students(school_id);
CREATE INDEX idx_students_admission_number ON students(admission_number);
CREATE INDEX idx_students_current_class_id ON students(current_class_id);

ALTER TABLE students ENABLE ROW LEVEL SECURITY;
CREATE POLICY students_tenant_isolation ON students
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Student Guardians/Parents (Many-to-Many)
CREATE TABLE student_guardians (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    guardian_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    relationship VARCHAR(100) NOT NULL, -- e.g., "father", "mother", "guardian"
    is_primary BOOLEAN DEFAULT false,
    can_pickup BOOLEAN DEFAULT true,
    can_receive_communication BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id, guardian_user_id)
);

CREATE INDEX idx_student_guardians_student_id ON student_guardians(student_id);
CREATE INDEX idx_student_guardians_guardian_user_id ON student_guardians(guardian_user_id);

-- Student Documents
CREATE TABLE student_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    document_type VARCHAR(100) NOT NULL, -- e.g., "birth_certificate", "photo", "medical_report"
    document_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size INT NOT NULL,
    mime_type VARCHAR(100),
    uploaded_by UUID REFERENCES users(id),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_student_documents_student_id ON student_documents(student_id);

-- ============================================================================
-- TEACHER MANAGEMENT
-- ============================================================================

-- Teachers
CREATE TABLE teachers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    employee_id VARCHAR(50) UNIQUE NOT NULL,
    department_id UUID REFERENCES departments(id),
    date_of_birth DATE NOT NULL,
    gender gender NOT NULL,
    date_of_joining DATE NOT NULL,
    qualification VARCHAR(255),
    specialization VARCHAR(255),
    experience_years INT DEFAULT 0,
    employment_type VARCHAR(50), -- e.g., "full_time", "part_time", "contract"
    salary DECIMAL(15, 2),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    emergency_contact_name VARCHAR(255),
    emergency_contact_phone VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, employee_id)
);

CREATE INDEX idx_teachers_tenant_id ON teachers(tenant_id);
CREATE INDEX idx_teachers_user_id ON teachers(user_id);
CREATE INDEX idx_teachers_school_id ON teachers(school_id);
CREATE INDEX idx_teachers_employee_id ON teachers(employee_id);
CREATE INDEX idx_teachers_department_id ON teachers(department_id);

ALTER TABLE teachers ENABLE ROW LEVEL SECURITY;
CREATE POLICY teachers_tenant_isolation ON teachers
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Teacher Subjects (Many-to-Many)
CREATE TABLE teacher_subjects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    class_id UUID REFERENCES classes(id),
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(teacher_id, subject_id, class_id, academic_year_id)
);

CREATE INDEX idx_teacher_subjects_teacher_id ON teacher_subjects(teacher_id);
CREATE INDEX idx_teacher_subjects_subject_id ON teacher_subjects(subject_id);
CREATE INDEX idx_teacher_subjects_class_id ON teacher_subjects(class_id);

-- ============================================================================
-- ENROLLMENT & CLASS ASSIGNMENTS
-- ============================================================================

-- Student Enrollments (Track year-by-year enrollment)
CREATE TABLE enrollments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    enrollment_date DATE NOT NULL DEFAULT CURRENT_DATE,
    roll_number VARCHAR(50),
    status VARCHAR(50) DEFAULT 'active', -- active, completed, withdrawn, transferred
    promoted_to_class_id UUID REFERENCES classes(id),
    final_result VARCHAR(50), -- passed, failed, promoted, detained
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id, class_id, academic_year_id)
);

CREATE INDEX idx_enrollments_tenant_id ON enrollments(tenant_id);
CREATE INDEX idx_enrollments_student_id ON enrollments(student_id);
CREATE INDEX idx_enrollments_class_id ON enrollments(class_id);
CREATE INDEX idx_enrollments_academic_year_id ON enrollments(academic_year_id);

ALTER TABLE enrollments ENABLE ROW LEVEL SECURITY;
CREATE POLICY enrollments_tenant_isolation ON enrollments
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Subject Enrollments (for universities/electives)
CREATE TABLE subject_enrollments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    enrollment_id UUID NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_id UUID REFERENCES teachers(id),
    enrolled_at DATE NOT NULL DEFAULT CURRENT_DATE,
    dropped_at DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(enrollment_id, subject_id)
);

CREATE INDEX idx_subject_enrollments_tenant_id ON subject_enrollments(tenant_id);
CREATE INDEX idx_subject_enrollments_enrollment_id ON subject_enrollments(enrollment_id);
CREATE INDEX idx_subject_enrollments_subject_id ON subject_enrollments(subject_id);

ALTER TABLE subject_enrollments ENABLE ROW LEVEL SECURITY;
CREATE POLICY subject_enrollments_tenant_isolation ON subject_enrollments
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- +goose Down
DROP TABLE IF EXISTS subject_enrollments CASCADE;
DROP TABLE IF EXISTS enrollments CASCADE;
DROP TABLE IF EXISTS teacher_subjects CASCADE;
DROP TABLE IF EXISTS teachers CASCADE;
DROP TABLE IF EXISTS student_documents CASCADE;
DROP TABLE IF EXISTS student_guardians CASCADE;
DROP TABLE IF EXISTS students CASCADE;
