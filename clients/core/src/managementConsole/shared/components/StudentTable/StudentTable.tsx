import {
  getStudentsWithCourses,
  type StudentWithCourses,
} from "@core/network/queries/getStudentsWithCourses";
import { type ColumnDef } from "@tanstack/react-table";
import { Role } from "@tumaet/prompt-shared-state";
import {
  PromptTable,
  type RowAction,
  type TableFilter,
} from "@tumaet/prompt-ui-components";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useStudentStore } from "../../store/student.store";
import { PrivacyDeletionInitiateDialog } from "../PrivacyDeletion/PrivacyDeletionInitiateDialog";
import { useHasRolePermission } from "../ShowForRole";
import { getStudentTableActions } from "./studentTableActions";
import { studentTableColumns } from "./studentTableColumns";
import { getStudentTableFilters } from "./studentTableFilters";

export const StudentTable = () => {
  const [studentsWithCourses, setStudentsWithCourses] = useState<
    Array<StudentWithCourses>
  >([]);
  const [studentsToDelete, setStudentsToDelete] = useState<string[] | null>(
    null,
  );

  const { upsertStudents } = useStudentStore();
  const isAdmin = useHasRolePermission({ roles: [Role.PROMPT_ADMIN] });

  const navigate = useNavigate();
  const openStudent = useCallback(
    (student: StudentWithCourses) =>
      navigate(`/management/students/${student.id}`),
    [navigate],
  );

  const fetchStudents = useCallback(async () => {
    const s = await getStudentsWithCourses();
    setStudentsWithCourses(s);
    upsertStudents(s);
  }, [upsertStudents]);

  useEffect(() => {
    fetchStudents();
  }, [fetchStudents]);

  const columns: ColumnDef<StudentWithCourses>[] = useMemo(
    () => studentTableColumns,
    [],
  );

  const filters: TableFilter[] = useMemo(
    () => getStudentTableFilters(studentsWithCourses),
    [studentsWithCourses],
  );

  const actions: RowAction<StudentWithCourses>[] = useMemo(
    () =>
      getStudentTableActions({
        openStudent,
        onInitiateDeletion: isAdmin
          ? (students) => setStudentsToDelete(students.map((s) => s.id))
          : undefined,
      }),
    [openStudent, isAdmin],
  );
  return (
    <div className="flex flex-col gap-3 w-full">
      <PromptTable
        data={studentsWithCourses}
        columns={columns}
        filters={filters}
        actions={actions}
        onRowClick={openStudent}
        initialState={{
          columnVisibility: {
            gender: false,
            studyDegree: false,
            nationality: false,
            fullname: false,
          },
        }}
      />
      <PrivacyDeletionInitiateDialog
        studentIDs={studentsToDelete ?? []}
        open={studentsToDelete !== null}
        onClose={() => {
          setStudentsToDelete(null);
          fetchStudents();
        }}
      />
    </div>
  );
};
