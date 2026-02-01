import { setActivePinia, createPinia } from 'pinia'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useStudentsStore } from './students'
import api from '@/services/api'

vi.mock('@/services/api')

describe('Students Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('starts with empty students list', () => {
      const store = useStudentsStore()
      expect(store.students).toEqual([])
    })

    it('starts with no selected student', () => {
      const store = useStudentsStore()
      expect(store.selectedStudent).toBeNull()
    })

    it('starts with no loading state', () => {
      const store = useStudentsStore()
      expect(store.loading).toBe(false)
    })

    it('starts with no error', () => {
      const store = useStudentsStore()
      expect(store.error).toBeNull()
    })
  })

  describe('Fetch Students', () => {
    it('fetches students successfully', async () => {
      const store = useStudentsStore()
      const mockStudents = [
        {
          id: '1',
          first_name: 'John',
          last_name: 'Doe',
          email: 'john@example.com',
          student_id: 'STU001',
          grade_level: '10',
          enrollment_date: '2024-01-01'
        },
        {
          id: '2',
          first_name: 'Jane',
          last_name: 'Smith',
          email: 'jane@example.com',
          student_id: 'STU002',
          grade_level: '11',
          enrollment_date: '2024-01-01'
        }
      ]

      vi.mocked(api.get).mockResolvedValue({ data: { students: mockStudents } })

      await store.fetchStudents()

      expect(store.students).toEqual(mockStudents)
      expect(store.loading).toBe(false)
      expect(store.error).toBeNull()
    })

    it('handles fetch error', async () => {
      const store = useStudentsStore()
      const errorMessage = 'Failed to fetch students'

      vi.mocked(api.get).mockRejectedValue({
        response: { data: { message: errorMessage } }
      })

      await store.fetchStudents()

      expect(store.students).toEqual([])
      expect(store.error).toBe(errorMessage)
      expect(store.loading).toBe(false)
    })

    it('sets loading state during fetch', async () => {
      const store = useStudentsStore()

      vi.mocked(api.get).mockImplementation(() => {
        expect(store.loading).toBe(true)
        return Promise.resolve({ data: { students: [] } })
      })

      await store.fetchStudents()
      expect(store.loading).toBe(false)
    })
  })

  describe('Get Student by ID', () => {
    it('fetches single student successfully', async () => {
      const store = useStudentsStore()
      const mockStudent = {
        id: '1',
        first_name: 'John',
        last_name: 'Doe',
        email: 'john@example.com',
        student_id: 'STU001',
        grade_level: '10',
        enrollment_date: '2024-01-01'
      }

      vi.mocked(api.get).mockResolvedValue({ data: mockStudent })

      await store.getStudent('1')

      expect(store.selectedStudent).toEqual(mockStudent)
    })

    it('handles get student error', async () => {
      const store = useStudentsStore()

      vi.mocked(api.get).mockRejectedValue(new Error('Student not found'))

      await store.getStudent('999')

      expect(store.selectedStudent).toBeNull()
      expect(store.error).toBeTruthy()
    })
  })

  describe('Create Student', () => {
    it('creates student successfully', async () => {
      const store = useStudentsStore()
      const newStudent = {
        first_name: 'New',
        last_name: 'Student',
        email: 'new@example.com',
        student_id: 'STU003',
        grade_level: '9'
      }

      const createdStudent = { ...newStudent, id: '3', enrollment_date: '2024-01-01' }

      vi.mocked(api.post).mockResolvedValue({ data: createdStudent })

      const result = await store.createStudent(newStudent)

      expect(result).toBe(true)
      expect(store.students).toContainEqual(createdStudent)
    })

    it('handles create error', async () => {
      const store = useStudentsStore()
      const newStudent = {
        first_name: 'Test',
        last_name: 'Student',
        email: 'invalid-email',
        student_id: 'STU004',
        grade_level: '9'
      }

      vi.mocked(api.post).mockRejectedValue({
        response: { data: { message: 'Invalid email' } }
      })

      const result = await store.createStudent(newStudent)

      expect(result).toBe(false)
      expect(store.error).toBe('Invalid email')
    })
  })

  describe('Update Student', () => {
    it('updates student successfully', async () => {
      const store = useStudentsStore()
      store.students = [
        {
          id: '1',
          first_name: 'John',
          last_name: 'Doe',
          email: 'john@example.com',
          student_id: 'STU001',
          grade_level: '10',
          enrollment_date: '2024-01-01'
        }
      ]

      const updates = { grade_level: '11' }
      const updatedStudent = { ...store.students[0], ...updates }

      vi.mocked(api.put).mockResolvedValue({ data: updatedStudent })

      const result = await store.updateStudent('1', updates)

      expect(result).toBe(true)
      expect(store.students[0].grade_level).toBe('11')
    })

    it('handles update error', async () => {
      const store = useStudentsStore()

      vi.mocked(api.put).mockRejectedValue(new Error('Update failed'))

      const result = await store.updateStudent('1', { grade_level: '12' })

      expect(result).toBe(false)
    })
  })

  describe('Delete Student', () => {
    it('deletes student successfully', async () => {
      const store = useStudentsStore()
      store.students = [
        {
          id: '1',
          first_name: 'John',
          last_name: 'Doe',
          email: 'john@example.com',
          student_id: 'STU001',
          grade_level: '10',
          enrollment_date: '2024-01-01'
        },
        {
          id: '2',
          first_name: 'Jane',
          last_name: 'Smith',
          email: 'jane@example.com',
          student_id: 'STU002',
          grade_level: '11',
          enrollment_date: '2024-01-01'
        }
      ]

      vi.mocked(api.delete).mockResolvedValue({})

      const result = await store.deleteStudent('1')

      expect(result).toBe(true)
      expect(store.students).toHaveLength(1)
      expect(store.students[0].id).toBe('2')
    })

    it('handles delete error', async () => {
      const store = useStudentsStore()

      vi.mocked(api.delete).mockRejectedValue(new Error('Delete failed'))

      const result = await store.deleteStudent('1')

      expect(result).toBe(false)
    })
  })

  describe('Search and Filter', () => {
    beforeEach(() => {
      const store = useStudentsStore()
      store.students = [
        {
          id: '1',
          first_name: 'John',
          last_name: 'Doe',
          email: 'john@example.com',
          student_id: 'STU001',
          grade_level: '10',
          enrollment_date: '2024-01-01'
        },
        {
          id: '2',
          first_name: 'Jane',
          last_name: 'Smith',
          email: 'jane@example.com',
          student_id: 'STU002',
          grade_level: '11',
          enrollment_date: '2024-01-01'
        },
        {
          id: '3',
          first_name: 'Bob',
          last_name: 'Johnson',
          email: 'bob@example.com',
          student_id: 'STU003',
          grade_level: '10',
          enrollment_date: '2024-01-01'
        }
      ]
    })

    it('filters students by grade level', () => {
      const store = useStudentsStore()
      const filtered = store.studentsByGrade('10')

      expect(filtered).toHaveLength(2)
      expect(filtered.map(s => s.id)).toEqual(['1', '3'])
    })

    it('searches students by name', () => {
      const store = useStudentsStore()
      const results = store.searchStudents('john')

      expect(results).toHaveLength(2) // John Doe and Bob Johnson
    })

    it('returns empty array for no matches', () => {
      const store = useStudentsStore()
      const results = store.searchStudents('nonexistent')

      expect(results).toEqual([])
    })
  })

  describe('Computed Properties', () => {
    it('counts total students', () => {
      const store = useStudentsStore()
      store.students = [
        { id: '1', first_name: 'A', last_name: 'B', email: 'a@b.com', student_id: 'S1', grade_level: '10', enrollment_date: '2024-01-01' },
        { id: '2', first_name: 'C', last_name: 'D', email: 'c@d.com', student_id: 'S2', grade_level: '11', enrollment_date: '2024-01-01' }
      ]

      expect(store.totalStudents).toBe(2)
    })

    it('checks if students are loaded', () => {
      const store = useStudentsStore()
      expect(store.hasStudents).toBe(false)

      store.students = [
        { id: '1', first_name: 'A', last_name: 'B', email: 'a@b.com', student_id: 'S1', grade_level: '10', enrollment_date: '2024-01-01' }
      ]

      expect(store.hasStudents).toBe(true)
    })
  })
})
