# AI-Friendly Repository Documentation

This directory contains minimal, high-signal documentation designed for AI assistants and smaller language models to effectively work with this codebase.

## ⚠️ CRITICAL: This is a SKELETON Application

**What this means:**
- Infrastructure is 100% implemented (database, cache, security, middleware)
- Database schema is 100% complete (15+ tables with multi-tenancy)
- Tests are comprehensive (6,800+ lines covering infrastructure)
- **BUT**: All business logic returns "to be implemented" placeholders
- The server runs, routes are registered, but NO CRUD operations work

**Start here:** Read `implementation-status.md` first to understand what exists vs. what's placeholder.

## Documentation Files

### 1. [implementation-status.md](./implementation-status.md) - **START HERE**
**What it covers:**
- Complete breakdown of implemented vs. placeholder code
- What actually works (infrastructure, security, schema)
- What doesn't work (all handlers, no repositories, no services)
- Next steps to make it functional

**Use this when:**
- You need to know if a feature is real or placeholder
- Planning what to implement next
- Understanding the 30% implemented / 70% TODO split

---

### 2. [repo-map.md](./repo-map.md)
**What it covers:**
- Repository purpose and tech stack
- Directory structure and module responsibilities
- 25 key choke points (HTTP entry, database, caching, auth)
- Dependency graph and configuration sources

**Use this when:**
- Getting initial orientation in the codebase
- Understanding the overall architecture
- Finding where specific functionality should live

---

### 3. [entrypoints.md](./entrypoints.md)
**What it covers:**
- API server startup and configuration
- All HTTP routes (with placeholder warnings)
- Makefile commands and CLI tools
- Health checks and metrics endpoints
- Frontend entrypoints

**Use this when:**
- Looking for specific API endpoints
- Understanding how to start the server
- Finding available commands (make run, make test, etc.)
- Checking route permissions

---

### 4. [components.md](./components.md)
**What it covers:**
- 10 major components with implementation status
- Infrastructure (✅ database, cache, events, config, observability)
- Security (✅ JWT, RBAC, crypto - fully implemented)
- Modules (❌ auth, users, academic - placeholders only)
- Component flows and interactions

**Use this when:**
- Understanding how a component works
- Finding where specific logic should go
- Tracing request/data flow through the system
- Looking for database/cache/event bus usage patterns

---

### 5. [data-and-contracts.md](./data-and-contracts.md)
**What it covers:**
- Complete PostgreSQL schema (15+ tables)
- Redis key patterns
- NATS event topics
- REST API contracts and error formats
- 30+ environment variables with defaults
- Missing tables (grades, attendance, etc.)

**Use this when:**
- Understanding database schema
- Finding table names and relationships
- Looking up environment variables
- Checking API request/response formats
- Understanding multi-tenancy (RLS policies)

---

### 6. [invariants-and-risk.md](./invariants-and-risk.md)
**What it covers:**
- 10 critical invariants (security, correctness, consistency)
- How to test each invariant
- Failure modes and recovery strategies
- "Do not break" list
- High coupling areas and blast radius

**Use this when:**
- Making changes to security code
- Understanding what must never break
- Testing multi-tenancy isolation
- Debugging auth/permission issues
- Assessing impact of changes

---

### 7. [run-and-test.md](./run-and-test.md)
**What it covers:**
- Build instructions (Go + Vue.js)
- Local development setup with Docker Compose
- How to run unit tests (2,560+ backend, 2,522+ frontend)
- Integration test setup (680+ lines)
- E2E tests (500+ lines with Playwright)
- Lint/format commands
- CI/CD configuration

**Use this when:**
- Setting up development environment
- Running tests
- Debugging build issues
- Understanding CI pipeline
- Formatting code

---

## Quick Start Guide

### For Understanding the Codebase
1. Read `implementation-status.md` - know what's real vs. placeholder
2. Skim `repo-map.md` - get oriented
3. Check `components.md` - understand the pieces

### For Implementing Features
1. Check `implementation-status.md` - see what needs implementing
2. Review `data-and-contracts.md` - understand schema and contracts
3. Read `invariants-and-risk.md` - know what not to break
4. Reference `components.md` - understand where code goes

### For Running/Testing
1. Follow `run-and-test.md` - set up environment
2. Check `entrypoints.md` - find commands
3. Verify with `implementation-status.md` - know what works

## File Size Summary

| File | Size | Purpose |
|------|------|---------|
| implementation-status.md | 13K | ⚠️ What's real vs. placeholder (READ FIRST) |
| data-and-contracts.md | 12K | Database schema, API contracts, config |
| invariants-and-risk.md | 14K | Security invariants, what not to break |
| run-and-test.md | 11K | Build, run, test instructions |
| components.md | 9.4K | Component architecture and flows |
| entrypoints.md | 7.3K | API routes, startup, commands |
| repo-map.md | 7.0K | High-level overview and choke points |
| **Total** | **73.7K** | Structured, factual, code-cited |

## Documentation Principles

All documentation follows these rules:
- ✅ **Factual**: Every claim includes file path citations
- ✅ **Concise**: Bullet points over prose
- ✅ **Prioritized**: 10-30 most important items per section
- ✅ **Honest**: Clear about placeholder vs. implemented
- ✅ **Actionable**: How to test, how to run, where to add code

## What to Expect

### This Documentation Will Tell You:
- ✅ Exact file paths and line numbers
- ✅ What's implemented with tests
- ✅ What's placeholder/stub
- ✅ How to run and test everything
- ✅ What will break the system
- ✅ Where to add new code

### This Documentation Will NOT:
- ❌ Claim features work when they don't
- ❌ Provide vague architectural theory
- ❌ Skip the hard truths about placeholders
- ❌ Give incomplete setup instructions
- ❌ Hide coupling or complexity

## Current State Summary

**Infrastructure (100% Complete):**
- PostgreSQL with connection pooling and multi-tenancy
- Redis with rate limiting
- NATS event bus
- MinIO/S3 storage
- JWT, RBAC, encryption
- Observability (logs, metrics, traces)
- 15+ database tables with RLS

**Business Logic (0% Complete):**
- Auth handlers: all return "to be implemented"
- User handlers: all return "to be implemented"
- Academic handlers: all return "to be implemented"
- No repository layer
- No service layer
- Frontend not connected

**Tests (100% Coverage for Implemented Code):**
- 6,800+ lines of tests
- Infrastructure: fully tested
- Security: fully tested
- Handlers: not tested (they're stubs)

## Next Actions Needed

To make this a working application, see `implementation-status.md` section "Next Steps to Make It Functional":

1. Implement Auth Module (register, login, logout)
2. Create Repository Layer (database queries)
3. Implement Users Module (CRUD operations)
4. Implement Academic Module (grades, attendance)
5. Connect Frontend (Vue.js to backend API)

---

**Last Updated:** 2026-02-08
**Documentation Version:** 1.1 (corrected for skeleton status)
**Repository State:** Architecture complete, business logic placeholder
