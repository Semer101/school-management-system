import api from "./axiosClient";
import type { APIResponse, PaginatedList } from "../types/api";
import type { Student, Teacher, Class, Subject } from "../types/academic";
import type { User, Role } from "../types/user";

// ── Users ────────────────────────────────────────────────
export const registerUser = (data: {
  name: string;
  email: string;
  password: string;
  role: Role;
  phone?: string;
}) => api.post<APIResponse<User>>("/api/admin/register", data);

// ── Students ─────────────────────────────────────────────
export const getStudents = (params?: {
  page?: number;
  page_size?: number;
  search?: string;
  stream?: string;
  grade_level?: string;
}) =>
  api.get<APIResponse<{ total: number; data: Student[] } | Student[]>>(
    "/api/admin/students",
    { params },
  );

export const getStudent = (id: number) =>
  api.get<APIResponse<Student>>(`/api/admin/students/${id}`);

export type CreateStudentPayload = {
  name: string;
  email: string;
  password: string;
  student_code?: string;
  class_id: number;
  parent_id: number;
  parent_name?: string;
  parent_email?: string;
  parent_phone?: string;
  date_of_birth?: string;
  stream: "Natural Science" | "Social Science" | "";
  grade_level: number;
};

export const createStudent = (data: CreateStudentPayload) =>
  api.post<APIResponse<Student>>("/api/admin/students", data);

export const updateStudent = (id: number, data: Partial<Student>) =>
  api.put<APIResponse<Student>>(`/api/admin/students/${id}`, data);

export const archiveStudent = (id: number) =>
  api.delete<APIResponse>(`/api/admin/students/${id}`);

// ── Teachers ─────────────────────────────────────────────
export const getTeachers = (params?: { page?: number; page_size?: number }) =>
  api.get<APIResponse<{ total: number; data: Teacher[] } | Teacher[]>>(
    "/api/admin/teachers",
    { params },
  );

export const getTeacher = (id: number) =>
  api.get<APIResponse<Teacher>>(`/api/admin/teachers/${id}`);

export type CreateTeacherPayload = {
  name: string;
  email: string;
  password: string;
  teacher_code: string;
  qualification?: string;
  phone?: string;
};

export const createTeacher = (data: CreateTeacherPayload) =>
  api.post<APIResponse<Teacher>>("/api/admin/teachers", data);

export const updateTeacher = (
  id: number,
  data: { qualification?: string; phone?: string },
) => api.put<APIResponse<Teacher>>(`/api/admin/teachers/${id}`, data);

export const archiveTeacher = (id: number) =>
  api.delete<APIResponse>(`/api/admin/teachers/${id}`);

// ── Classes ──────────────────────────────────────────────
export const getClasses = (params?: { page?: number; page_size?: number }) =>
  api.get<APIResponse<{ total: number; data: Class[] } | Class[]>>(
    "/api/admin/classes",
    { params },
  );

export const createClass = (data: {
  name?: string;
  grade_level: number;
  section: string;
  stream?: string;
  status?: string;
  year: number;
  teacher_id: number;
}) => api.post<APIResponse<Class>>("/api/admin/classes", data);

export interface ChildInfo {
  id: number;
  name: string;
  student_code: string;
  grade: string;
  section: string;
}

export interface ParentRow {
  id: number;
  name: string;
  email: string;
  phone: string;
  avatar_url?: string;
  is_active: boolean;
  status: string;
  children_count: number;
  children: ChildInfo[];
  student_names: string;
}

export interface AdminRow {
  id: number;
  name: string;
  email: string;
  phone: string;
  avatar_url?: string;
  is_active: boolean;
  status: string;
}

export const getAdmins = (params?: { page?: number }) =>
  api.get<APIResponse<{ total: number; data: AdminRow[] }>>(
    "/api/admin/admins",
    { params },
  );

export const updateAdmin = (
  id: number,
  data: { name: string; email: string; phone?: string },
) => api.put<APIResponse>(`/api/admin/admins/${id}`, data);

export const archiveAdmin = (id: number) =>
  api.delete<APIResponse>(`/api/admin/admins/${id}`);

export const getParents = (params?: { page?: number; search?: string }) =>
  api.get<APIResponse<{ total: number; data: ParentRow[] }>>(
    "/api/admin/parents",
    { params },
  );

export const updateParent = (
  id: number,
  data: { name: string; email: string; phone?: string },
) => api.put<APIResponse>(`/api/admin/parents/${id}`, data);

export const archiveParent = (id: number) =>
  api.delete<APIResponse>(`/api/admin/parents/${id}`);

export const updateClass = (id: number, data: Partial<Class>) =>
  api.put<APIResponse<Class>>(`/api/admin/classes/${id}`, data);

export const archiveClass = (id: number) =>
  api.delete<APIResponse>(`/api/admin/classes/${id}`);

// ── Subjects ─────────────────────────────────────────────
export const getSubjects = (params?: {
  page?: number;
  page_size?: number;
  stream?: string;
  grade_level?: string;
}) =>
  api.get<APIResponse<{ total: number; data: Subject[] } | Subject[]>>(
    "/api/admin/subjects",
    { params },
  );

export const createSubject = (data: {
  name: string;
  code: string;
  grade_level?: number;
  stream?: string;
  status?: string;
  teacher_id: number;
}) => api.post<APIResponse<Subject>>("/api/admin/subjects", data);

export const updateSubject = (id: number, data: Partial<Subject>) =>
  api.put<APIResponse<Subject>>(`/api/admin/subjects/${id}`, data);

export const archiveSubject = (id: number) =>
  api.delete<APIResponse>(`/api/admin/subjects/${id}`);

// ── Enrollment ───────────────────────────────────────────
export const enrollStudent = (student_id: number, subject_id: number) =>
  api.post<APIResponse>("/api/admin/enroll", { student_id, subject_id });

export const unenrollStudent = (student_id: number, subject_id: number) =>
  api.delete<APIResponse>("/api/admin/unenroll", {
    data: { student_id, subject_id },
  });

// ── Attendance Summary ───────────────────────────────────
export type AttendanceSummaryRow = {
  student_name: string;
  student_code: string;
  class_name: string;
  grade_level: number;
  section: string;
  date: string;
  status: string;
};

export const getAttendanceSummary = (params?: {
  date?: string;
  grade_level?: string;
  section?: string;
  class_id?: string;
}) =>
  api.get<
    APIResponse<AttendanceSummaryRow[] | PaginatedList<AttendanceSummaryRow>>
  >("/api/admin/attendance/summary", { params });

// ── Admin Locker ─────────────────────────────────────────
export const adminGetLockerFiles = (studentID: number) =>
  api.get<APIResponse>(`/api/admin/locker/student/${studentID}`);

// ── Notifications ────────────────────────────────────────
export const broadcastAnnouncement = (data: {
  title: string;
  body: string;
  target_roles: string[];
}) => api.post<APIResponse>("/api/admin/notify/broadcast", data);

export const notifyAbsentParents = () =>
  api.post<APIResponse>("/api/admin/notify/absences");

export type EnrollmentStatusRow = {
  subject_id: number;
  subject_name: string;
  subject_code: string;
  enrolled: boolean;
};

export const getStudentEnrollmentStatus = (studentId: number) =>
  api.get<APIResponse<EnrollmentStatusRow[]>>(
    `/api/admin/students/${studentId}/enrollment-status`,
  );

export const promoteStudent = (studentId: number) =>
  api.post<APIResponse>(`/api/admin/students/${studentId}/promote`);

export const getPromotionPreview = (studentId: number) =>
  api.get<
    APIResponse<{
      promotion_status: string;
      failed_subjects: number;
      can_promote: boolean;
    }>
  >(`/api/admin/students/${studentId}/promotion-preview`);

export const uploadUserAvatar = (userId: number, file: File) => {
  const form = new FormData();
  form.append("avatar", file);
  return api.post<{ data: { avatar_url: string } }>(
    `/api/admin/users/${userId}/avatar`,
    form,
    { headers: { "Content-Type": "multipart/form-data" } },
  );
};
