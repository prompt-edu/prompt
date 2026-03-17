import { Student } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'

interface studentById {
  [id: string]: Student
}

interface StudentStore {
  studentsById: studentById
  students: Student[]

  upsertStudent: (student: Student) => void
  upsertStudents: (students: Student[]) => void
}

export const useStudentStore = create<StudentStore>((set) => ({
  studentsById: {},
  students: [],
  upsertStudent: (student) =>
    set((state) => {
      const studentsById = { ...state.studentsById, [student.id!]: student }
      const students = state.students.map((s) => (s.id === student.id ? student : s))
      if (!state.studentsById[student.id!]) {
        students.push(student)
      }
      return { studentsById, students }
    }),

  upsertStudents: (students) =>
    set((state) => {
      const studentsById = { ...state.studentsById }
      let nextStudents = state.students
      students.forEach((student) => {
        const exists = !!studentsById[student.id!]
        studentsById[student.id!] = student
        nextStudents = nextStudents.map((s) => (s.id === student.id ? student : s))
        if (!exists) nextStudents = [...nextStudents, student]
      })
      return { studentsById, students: nextStudents }
    }),
}))
