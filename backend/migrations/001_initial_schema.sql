-- +goose Up
-- OpenSMS Initial Database Schema with Multi-Tenancy Support
-- PostgreSQL 16+ with Row-Level Security

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create custom types
CREATE TYPE user_role AS ENUM (
    'super_admin', 'school_admin', 'principal', 'teacher',
    'student', 'parent', 'librarian', 'accountant',
    'registrar', 'counselor', 'nurse', 'transport_admin', 'hostel_warden'
);

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'deleted');
CREATE TYPE gender AS ENUM ('male', 'female', 'other', 'prefer_not_to_say');
CREATE TYPE academic_level AS ENUM ('primary', 'secondary', 'tertiary');
CREATE TYPE payment_status AS ENUM ('pending', 'partial', 'paid', 'overdue', 'cancelled');
CREATE TYPE attendance_status AS ENUM ('present', 'absent', 'late', 'excused', 'sick_leave');

-- ============================================================================
-- TENANT & ORGANIZATION TABLES
-- ============================================================================

-- Tenants (Schools/Institutions) - Top-level isolation
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255) UNIQUE,
    academic_level academic_level NOT NULL,
    timezone VARCHAR(100) DEFAULT 'UTC',
    country VARCHAR(2) NOT NULL, -- ISO 3166-1 alpha-2
    settings JSONB DEFAULT '{}',
    subscription_plan VARCHAR(50) DEFAULT 'free',
    subscription_expires_at TIMESTAMP,
    max_students INT DEFAULT 100,
    max_teachers INT DEFAULT 10,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_domain ON tenants(domain);
CREATE INDEX idx_tenants_is_active ON tenants(is_active);

-- Schools (for multi-campus support within a tenant)
CREATE TABLE schools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    phone VARCHAR(50),
    email VARCHAR(255),
    principal_id UUID, -- FK added later
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_schools_tenant_id ON schools(tenant_id);
CREATE INDEX idx_schools_code ON schools(code);

-- Enable RLS
ALTER TABLE schools ENABLE ROW LEVEL SECURITY;

CREATE POLICY schools_tenant_isolation ON schools
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- ============================================================================
-- USER MANAGEMENT TABLES
-- ============================================================================

-- Users - Core authentication table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    status user_status DEFAULT 'active',
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    phone VARCHAR(50),
    avatar_url TEXT,
    last_login_at TIMESTAMP,
    last_login_ip INET,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    email_verified BOOLEAN DEFAULT false,
    email_verification_token VARCHAR(255),
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP,
    must_change_password BOOLEAN DEFAULT false,
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret VARCHAR(255),
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- Enable RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Refresh Tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    replaced_by_token VARCHAR(255)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- ============================================================================
-- ACADEMIC STRUCTURE TABLES
-- ============================================================================

-- Academic Years/Sessions
CREATE TABLE academic_years (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    school_id UUID REFERENCES schools(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_current BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, school_id, name)
);

CREATE INDEX idx_academic_years_tenant_id ON academic_years(tenant_id);
CREATE INDEX idx_academic_years_school_id ON academic_years(school_id);
CREATE INDEX idx_academic_years_is_current ON academic_years(is_current);

ALTER TABLE academic_years ENABLE ROW LEVEL SECURITY;
CREATE POLICY academic_years_tenant_isolation ON academic_years
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Terms/Semesters
CREATE TABLE terms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    term_number INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_current BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_terms_tenant_id ON terms(tenant_id);
CREATE INDEX idx_terms_academic_year_id ON terms(academic_year_id);
CREATE INDEX idx_terms_is_current ON terms(is_current);

ALTER TABLE terms ENABLE ROW LEVEL SECURITY;
CREATE POLICY terms_tenant_isolation ON terms
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Grade Levels (Flexible for all education levels)
CREATE TABLE grade_levels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    school_id UUID REFERENCES schools(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- e.g., "Grade 1", "Form 4", "Year 1", "100 Level"
    level_order INT NOT NULL, -- For sorting
    academic_level academic_level NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, school_id, name)
);

CREATE INDEX idx_grade_levels_tenant_id ON grade_levels(tenant_id);
CREATE INDEX idx_grade_levels_school_id ON grade_levels(school_id);

ALTER TABLE grade_levels ENABLE ROW LEVEL SECURITY;
CREATE POLICY grade_levels_tenant_isolation ON grade_levels
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Departments/Faculties
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    head_id UUID, -- FK to users
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, school_id, code)
);

CREATE INDEX idx_departments_tenant_id ON departments(tenant_id);
CREATE INDEX idx_departments_school_id ON departments(school_id);

ALTER TABLE departments ENABLE ROW LEVEL SECURITY;
CREATE POLICY departments_tenant_isolation ON departments
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Classes/Sections
CREATE TABLE classes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    grade_level_id UUID NOT NULL REFERENCES grade_levels(id) ON DELETE CASCADE,
    department_id UUID REFERENCES departments(id),
    name VARCHAR(255) NOT NULL,
    section VARCHAR(50), -- e.g., "A", "B", "Science"
    class_teacher_id UUID, -- FK to users
    capacity INT DEFAULT 30,
    room_number VARCHAR(50),
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, school_id, grade_level_id, section, academic_year_id)
);

CREATE INDEX idx_classes_tenant_id ON classes(tenant_id);
CREATE INDEX idx_classes_school_id ON classes(school_id);
CREATE INDEX idx_classes_grade_level_id ON classes(grade_level_id);
CREATE INDEX idx_classes_academic_year_id ON classes(academic_year_id);

ALTER TABLE classes ENABLE ROW LEVEL SECURITY;
CREATE POLICY classes_tenant_isolation ON classes
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- Subjects/Courses
CREATE TABLE subjects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    department_id UUID REFERENCES departments(id),
    credit_hours INT DEFAULT 0,
    is_elective BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, school_id, code)
);

CREATE INDEX idx_subjects_tenant_id ON subjects(tenant_id);
CREATE INDEX idx_subjects_school_id ON subjects(school_id);
CREATE INDEX idx_subjects_code ON subjects(code);

ALTER TABLE subjects ENABLE ROW LEVEL SECURITY;
CREATE POLICY subjects_tenant_isolation ON subjects
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- +goose Down
DROP TABLE IF EXISTS subjects CASCADE;
DROP TABLE IF EXISTS classes CASCADE;
DROP TABLE IF EXISTS departments CASCADE;
DROP TABLE IF EXISTS grade_levels CASCADE;
DROP TABLE IF EXISTS terms CASCADE;
DROP TABLE IF EXISTS academic_years CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS schools CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

DROP TYPE IF EXISTS attendance_status;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS academic_level;
DROP TYPE IF EXISTS gender;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;

DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
