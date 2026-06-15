import { useRef, useState, useEffect, useMemo, useReducer } from "react";
import { useAuth } from "../hooks/useAuth";
import { useRole } from "../hooks/useRole";
import { Button } from "../components/ui/Button";
import { PageHeader } from "../components/ui/PageHeader";
import { QRCodeSVG } from "qrcode.react";
import { Printer, Download, Search, X, Camera } from "lucide-react";
import {
  getStudents,
  getTeachers,
  getParents,
  getAdmins,
  uploadUserAvatar,
} from "../api/admin";
import type { Student, Teacher } from "../types/academic";
import type { ParentRow, AdminRow } from "../api/admin";
import type { Role } from "../types/user";
import { listFromApi } from "../types/api";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";
const VALID_UNTIL_YEAR = 2027;
const SCHOOL_NAME = "SMS Portal";

type IdCardRole = Role;

interface CardData {
  id: number;
  name: string;
  email: string;
  role: IdCardRole;
  code?: string;
  avatar_url?: string;
  phone?: string;
  // Student-specific
  grade_level?: number;
  section?: string;
  stream?: string;
  gender?: string;
  // Teacher-specific
  department?: string;
  qualification?: string;
  // Parent-specific
  children?: {
    id: number;
    name: string;
    student_code: string;
    grade: string;
    section: string;
  }[];
  // Admin-specific
  position?: string;
}

// ─── Utility to build a profile image URL ────────────────────────────────────
function profileUrl(avatar_url?: string): string | null {
  return avatar_url && avatar_url.length > 0
    ? `${API_BASE}${avatar_url}`
    : null;
}

// ─── Helper to build the QR data string ─────────────────────────────────────
// Each role encodes all fields relevant for identity verification.
function qrDataFromCard(card: CardData): string {
  const base = {
    school: SCHOOL_NAME,
    valid_until: VALID_UNTIL_YEAR,
    role: card.role,
    id: card.id,
    code: card.code ?? "",
    name: card.name,
    email: card.email,
    phone: card.phone ?? "",
  };

  switch (card.role) {
    case "Student":
      return JSON.stringify({
        ...base,
        grade: card.grade_level ?? "",
        section: card.section ?? "",
        stream: card.stream ?? "",
      });
    case "Teacher":
      return JSON.stringify({
        ...base,
        department: card.department ?? "",
        qualification: card.qualification ?? "",
      });
    case "Parent":
      return JSON.stringify({
        ...base,
        children:
          card.children
            ?.map(
              (c) =>
                `${c.name} (${c.grade}${c.section ? ` ${c.section}` : ""})`,
            )
            .join(", ") ?? "",
      });
    case "Admin":
      return JSON.stringify({
        ...base,
        position: card.position ?? "",
      });
    default:
      return JSON.stringify(base);
  }
}

// ─── School Logo Component ──────────────────────────────────────────────────
function SchoolLogo({ size = "md" }: { size?: "sm" | "md" }) {
  const dim = size === "sm" ? "w-7 h-7" : "w-8 h-8";
  return (
    <div
      className={`${dim} rounded-lg bg-white/20 border border-white/30 flex items-center justify-center`}
    >
      <span className="text-white font-mono text-sm font-bold">S</span>
    </div>
  );
}

// ─── Role-specific ID Card components ────────────────────────────────────────

function StudentIdCard({ card }: { card: CardData }) {
  const imgUrl = profileUrl(card.avatar_url);
  const qrData = qrDataFromCard(card);

  return (
    <div
      id={`card-${card.role}-${card.id}`}
      className="id-card relative w-[300px] bg-white rounded-2xl shadow-xl border border-gray-200 overflow-hidden"
    >
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-800 px-3 py-3 text-white text-center">
        <div className="flex items-center justify-center gap-1.5 mb-0.5">
          <SchoolLogo size="sm" />
          <p className="text-[11px] font-bold tracking-wide">{SCHOOL_NAME}</p>
        </div>
        <p className="text-[9px] uppercase tracking-widest opacity-80">
          Student ID Card
        </p>
      </div>

      {/* Photo */}
      <div className="flex justify-center pt-4 -mt-2">
        {imgUrl ? (
          <img
            src={imgUrl}
            alt={card.name}
            className="w-16 h-16 rounded-full border-3 border-white shadow-sm object-cover bg-gray-100"
          />
        ) : (
          <div className="w-16 h-16 rounded-full border-3 border-white shadow-sm bg-blue-100 flex items-center justify-center text-blue-600 text-xl font-bold">
            {card.name[0]?.toUpperCase()}
          </div>
        )}
      </div>

      {/* Details */}
      <div className="px-3 py-2 text-center space-y-1">
        <h3 className="text-sm font-bold text-gray-900 break-words">
          {card.name}
        </h3>
        {card.code && (
          <p className="text-[10px] font-mono text-gray-500">ID: {card.code}</p>
        )}
        {card.grade_level && (
          <p className="text-[10px] text-gray-500">
            Grade {card.grade_level}
            {card.section ? ` — ${card.section}` : ""}
            {card.stream ? ` (${card.stream})` : ""}
          </p>
        )}
        {card.gender && (
          <p className="text-[10px] text-gray-400">{card.gender}</p>
        )}
      </div>

      {/* QR */}
      <div className="flex justify-center pb-3">
        <div className="bg-white p-1 rounded-lg border border-gray-200">
          <QRCodeSVG value={qrData} size={80} level="M" />
        </div>
      </div>

      {/* Footer */}
      <div className="bg-gray-50 px-3 py-1.5 text-center border-t border-gray-200">
        <p className="text-[8px] uppercase tracking-widest text-gray-400">
          {SCHOOL_NAME} • Valid: {VALID_UNTIL_YEAR}
        </p>
      </div>
    </div>
  );
}

function TeacherIdCard({ card }: { card: CardData }) {
  const imgUrl = profileUrl(card.avatar_url);
  const qrData = qrDataFromCard(card);

  return (
    <div
      id={`card-${card.role}-${card.id}`}
      className="id-card relative w-[300px] bg-white rounded-2xl shadow-xl border border-gray-200 overflow-hidden"
    >
      <div className="bg-gradient-to-r from-emerald-600 to-emerald-800 px-3 py-3 text-white text-center">
        <div className="flex items-center justify-center gap-1.5 mb-0.5">
          <SchoolLogo size="sm" />
          <p className="text-[11px] font-bold tracking-wide">{SCHOOL_NAME}</p>
        </div>
        <p className="text-[9px] uppercase tracking-widest opacity-80">
          Staff ID Card
        </p>
      </div>

      <div className="flex justify-center pt-4 -mt-2">
        {imgUrl ? (
          <img
            src={imgUrl}
            alt={card.name}
            className="w-16 h-16 rounded-full border-3 border-white shadow-sm object-cover bg-gray-100"
          />
        ) : (
          <div className="w-16 h-16 rounded-full border-3 border-white shadow-sm bg-emerald-100 flex items-center justify-center text-emerald-600 text-xl font-bold">
            {card.name[0]?.toUpperCase()}
          </div>
        )}
      </div>

      <div className="px-3 py-2 text-center space-y-1">
        <h3 className="text-sm font-bold text-gray-900 break-words">
          {card.name}
        </h3>
        {card.code && (
          <p className="text-[10px] font-mono text-gray-500">
            Staff ID: {card.code}
          </p>
        )}
        {card.department && (
          <p className="text-[10px] text-gray-500">{card.department}</p>
        )}
        {card.phone && (
          <p className="text-[10px] text-gray-400">{card.phone}</p>
        )}
      </div>

      <div className="flex justify-center pb-3">
        <div className="bg-white p-1 rounded-lg border border-gray-200">
          <QRCodeSVG value={qrData} size={80} level="M" />
        </div>
      </div>

      <div className="bg-gray-50 px-3 py-1.5 text-center border-t border-gray-200">
        <p className="text-[8px] uppercase tracking-widest text-gray-400">
          {SCHOOL_NAME} • Valid: {VALID_UNTIL_YEAR}
        </p>
      </div>
    </div>
  );
}

function ParentIdCard({ card }: { card: CardData }) {
  const imgUrl = profileUrl(card.avatar_url);
  const qrData = qrDataFromCard(card);

  return (
    <div
      id={`card-${card.role}-${card.id}`}
      className="id-card relative w-[300px] bg-white rounded-2xl shadow-xl border border-gray-200 overflow-hidden"
    >
      <div className="bg-gradient-to-r from-purple-600 to-purple-800 px-3 py-3 text-white text-center">
        <div className="flex items-center justify-center gap-1.5 mb-0.5">
          <SchoolLogo size="sm" />
          <p className="text-[11px] font-bold tracking-wide">{SCHOOL_NAME}</p>
        </div>
        <p className="text-[9px] uppercase tracking-widest opacity-80">
          Parent ID Card
        </p>
      </div>

      <div className="flex justify-center pt-4 -mt-2">
        {imgUrl ? (
          <img
            src={imgUrl}
            alt={card.name}
            className="w-16 h-16 rounded-full border-3 border-white shadow-sm object-cover bg-gray-100"
          />
        ) : (
          <div className="w-16 h-16 rounded-full border-3 border-white shadow-sm bg-purple-100 flex items-center justify-center text-purple-600 text-xl font-bold">
            {card.name[0]?.toUpperCase()}
          </div>
        )}
      </div>

      <div className="px-3 py-2 text-center space-y-1">
        <h3 className="text-sm font-bold text-gray-900 break-words">
          {card.name}
        </h3>
        {card.code && (
          <p className="text-[10px] font-mono text-gray-500">
            Parent ID: {card.code}
          </p>
        )}
        {card.phone && (
          <p className="text-[10px] text-gray-500">{card.phone}</p>
        )}
        {card.children && card.children.length > 0 && (
          <div className="mt-1 pt-1.5 border-t border-gray-100">
            <p className="text-[9px] uppercase tracking-wider text-gray-400 mb-1">
              Student(s)
            </p>
            {card.children.map((child) => (
              <p key={child.id} className="text-[10px] text-gray-600">
                {child.name}{" "}
                {child.grade
                  ? `• ${child.grade}${child.section ? ` (${child.section})` : ""}`
                  : ""}
              </p>
            ))}
          </div>
        )}
      </div>

      <div className="flex justify-center pb-3">
        <div className="bg-white p-1 rounded-lg border border-gray-200">
          <QRCodeSVG value={qrData} size={80} level="M" />
        </div>
      </div>

      <div className="bg-gray-50 px-3 py-1.5 text-center border-t border-gray-200">
        <p className="text-[8px] uppercase tracking-widest text-gray-400">
          {SCHOOL_NAME} • Valid: {VALID_UNTIL_YEAR}
        </p>
      </div>
    </div>
  );
}

function AdminIdCard({ card }: { card: CardData }) {
  const imgUrl = profileUrl(card.avatar_url);
  const qrData = qrDataFromCard(card);

  return (
    <div
      id={`card-${card.role}-${card.id}`}
      className="id-card relative w-[300px] bg-white rounded-2xl shadow-xl border border-gray-200 overflow-hidden"
    >
      <div className="bg-gradient-to-r from-amber-600 to-amber-800 px-3 py-3 text-white text-center">
        <div className="flex items-center justify-center gap-1.5 mb-0.5">
          <SchoolLogo size="sm" />
          <p className="text-[11px] font-bold tracking-wide">{SCHOOL_NAME}</p>
        </div>
        <p className="text-[9px] uppercase tracking-widest opacity-80">
          Admin ID Card
        </p>
      </div>

      <div className="flex justify-center pt-4 -mt-2">
        {imgUrl ? (
          <img
            src={imgUrl}
            alt={card.name}
            className="w-16 h-16 rounded-full border-3 border-white shadow-sm object-cover bg-gray-100"
          />
        ) : (
          <div className="w-16 h-16 rounded-full border-3 border-white shadow-sm bg-amber-100 flex items-center justify-center text-amber-600 text-xl font-bold">
            {card.name[0]?.toUpperCase()}
          </div>
        )}
      </div>

      <div className="px-3 py-2 text-center space-y-1">
        <h3 className="text-sm font-bold text-gray-900 break-words">
          {card.name}
        </h3>
        {card.code && (
          <p className="text-[10px] font-mono text-gray-500">
            Admin ID: {card.code}
          </p>
        )}
        {card.position && (
          <p className="text-[10px] text-gray-500">{card.position}</p>
        )}
        {card.phone && (
          <p className="text-[10px] text-gray-400">{card.phone}</p>
        )}
      </div>

      <div className="flex justify-center pb-3">
        <div className="bg-white p-1 rounded-lg border border-gray-200">
          <QRCodeSVG value={qrData} size={80} level="M" />
        </div>
      </div>

      <div className="bg-gray-50 px-3 py-1.5 text-center border-t border-gray-200">
        <p className="text-[8px] uppercase tracking-widest text-gray-400">
          {SCHOOL_NAME} • Valid: {VALID_UNTIL_YEAR}
        </p>
      </div>
    </div>
  );
}

// ─── Map raw API data to CardData ──────────────────────────────────────────

function studentToCardData(s: Student): CardData {
  return {
    id: s.user?.id ?? s.id,
    name: s.user?.name ?? "Unknown",
    email: s.user?.email ?? "",
    role: "Student",
    code: s.student_code,
    avatar_url: s.user?.avatar_url,
    phone: s.user?.phone,
    grade_level: s.grade_level,
    section: s.class?.section,
    stream: s.stream,
  };
}

function teacherToCardData(t: Teacher): CardData {
  return {
    id: t.user?.id ?? t.id,
    name: t.user?.name ?? "Unknown",
    email: t.user?.email ?? "",
    role: "Teacher",
    code: t.teacher_code,
    avatar_url: t.user?.avatar_url,
    phone: t.user?.phone,
    department: t.department,
    qualification: t.qualification,
  };
}

function parentRowToCardData(p: ParentRow): CardData {
  return {
    id: p.id,
    name: p.name,
    email: p.email,
    role: "Parent",
    code: `PRT-${p.id}`,
    avatar_url: p.avatar_url,
    phone: p.phone,
    children: p.children?.map((c) => ({
      id: c.id,
      name: c.name,
      student_code: c.student_code,
      grade: c.grade,
      section: c.section,
    })),
  };
}

function adminRowToCardData(a: AdminRow): CardData {
  return {
    id: a.id,
    name: a.name,
    email: a.email,
    role: "Admin",
    code: `ADM-${a.id}`,
    avatar_url: a.avatar_url,
    phone: a.phone,
    position: "Administrator",
  };
}

// ─── AdminPhotoOverlay ───────────────────────────────────────────────────

function AdminPhotoOverlay({
  card,
  onUploaded,
}: {
  card: CardData;
  onUploaded: (avatarUrl: string) => void;
}) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);

  const handleFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const res = await uploadUserAvatar(card.id, file);
      const url = res.data?.data?.avatar_url ?? "";
      if (url) onUploaded(url);
    } catch {
      // silently ignore — user can retry
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = "";
    }
  };

  return (
    <div className="relative inline-block">
      <IdCardRenderer card={card} />
      {/* Camera button floats over the avatar area */}
      <button
        type="button"
        onClick={() => fileRef.current?.click()}
        disabled={uploading}
        title="Change photo"
        className="absolute top-[68px] left-1/2 -translate-x-1/2 w-7 h-7 rounded-full bg-black/60 hover:bg-black/80 text-white flex items-center justify-center shadow transition-opacity opacity-0 hover:opacity-100 focus:opacity-100 disabled:opacity-50"
        style={{ pointerEvents: "auto" }}
      >
        <Camera className="w-3.5 h-3.5" />
      </button>
      {uploading && (
        <div className="absolute top-[68px] left-1/2 -translate-x-1/2 text-[10px] bg-black/70 text-white rounded px-1.5 py-0.5 mt-8 whitespace-nowrap">
          Uploading…
        </div>
      )}
      <input
        ref={fileRef}
        type="file"
        accept="image/jpeg,image/png,image/webp"
        className="hidden"
        onChange={handleFile}
      />
      <p className="text-center text-[10px] text-muted mt-1.5">
        Hover photo to change
      </p>
    </div>
  );
}

// ─── IdCardRenderer ──────────────────────────────────────────────────────

function IdCardRenderer({ card }: { card: CardData }) {
  switch (card.role) {
    case "Student":
      return <StudentIdCard card={card} />;
    case "Teacher":
      return <TeacherIdCard card={card} />;
    case "Parent":
      return <ParentIdCard card={card} />;
    case "Admin":
    default:
      return <AdminIdCard card={card} />;
  }
}

// ─── Main Page ──────────────────────────────────────────────────────────────

export default function IdCardPage() {
  const { user: authUser } = useAuth();
  const { role } = useRole();
  const printRef = useRef<HTMLDivElement>(null);

  // Role selection state
  const [selectedRole, setSelectedRole] = useState<IdCardRole | "">("");
  const [batchMode, setBatchMode] = useState<"single" | "batch">("single");
  const [selectedUsers, setSelectedUsers] = useState<CardData[]>([]);

  type UsersState = {
    users: CardData[];
    searchQuery: string;
    selectedUser: CardData | null;
    loadingUsers: boolean;
  };

  type UsersAction =
    | { type: "RESET" }
    | { type: "START_LOADING" }
    | { type: "LOADED_USERS"; users: CardData[] }
    | { type: "LOAD_ERROR" }
    | { type: "SET_SEARCH"; query: string }
    | { type: "SELECT_USER"; user: CardData | null }
    | {
        type: "UPDATE_USER_AVATAR";
        userId: number;
        role: IdCardRole;
        avatarUrl: string;
      };

  function usersReducer(state: UsersState, action: UsersAction): UsersState {
    switch (action.type) {
      case "RESET":
        return {
          users: [],
          searchQuery: "",
          selectedUser: null,
          loadingUsers: false,
        };
      case "START_LOADING":
        return {
          ...state,
          loadingUsers: true,
          selectedUser: null,
          searchQuery: "",
        };
      case "LOADED_USERS":
        return { ...state, users: action.users, loadingUsers: false };
      case "LOAD_ERROR":
        return { ...state, users: [], loadingUsers: false };
      case "SET_SEARCH":
        return { ...state, searchQuery: action.query };
      case "SELECT_USER":
        return { ...state, selectedUser: action.user };
      case "UPDATE_USER_AVATAR":
        return {
          ...state,
          users: state.users.map((u) =>
            u.id === action.userId && u.role === action.role
              ? { ...u, avatar_url: action.avatarUrl }
              : u,
          ),
          selectedUser:
            state.selectedUser?.id === action.userId &&
            state.selectedUser?.role === action.role
              ? { ...state.selectedUser, avatar_url: action.avatarUrl }
              : state.selectedUser,
        };
    }
  }

  const [usersState, dispatch] = useReducer(usersReducer, {
    users: [],
    searchQuery: "",
    selectedUser: null,
    loadingUsers: false,
  });
  const { users, searchQuery, selectedUser, loadingUsers } = usersState;

  // Filter users by search query
  const filteredUsers = useMemo(() => {
    if (!searchQuery.trim()) return users;
    const q = searchQuery.toLowerCase();
    return users.filter(
      (u) =>
        u.name.toLowerCase().includes(q) ||
        u.email.toLowerCase().includes(q) ||
        (u.code && u.code.toLowerCase().includes(q)),
    );
  }, [users, searchQuery]);

  // Load users when role changes
  useEffect(() => {
    if (!selectedRole) return;

    const loadUsers = async () => {
      try {
        switch (selectedRole) {
          case "Student": {
            const res = await getStudents({ page_size: 200 });
            const data = listFromApi(res.data) as Student[];
            dispatch({
              type: "LOADED_USERS",
              users: data.map(studentToCardData),
            });
            break;
          }
          case "Teacher": {
            const res = await getTeachers({ page_size: 200 });
            const data = listFromApi(res.data) as Teacher[];
            dispatch({
              type: "LOADED_USERS",
              users: data.map(teacherToCardData),
            });
            break;
          }
          case "Parent": {
            const res = await getParents({ page: 1 });
            const payload = res.data.data;
            const data = (
              Array.isArray(payload)
                ? payload
                : ((payload as { data: ParentRow[] })?.data ?? [])
            ) as ParentRow[];
            dispatch({
              type: "LOADED_USERS",
              users: data.map(parentRowToCardData),
            });
            break;
          }
          case "Admin": {
            const res = await getAdmins({ page: 1 });
            const payload = res.data.data;
            const data = (
              Array.isArray(payload)
                ? payload
                : ((payload as { data: AdminRow[] })?.data ?? [])
            ) as AdminRow[];
            dispatch({
              type: "LOADED_USERS",
              users: data.map(adminRowToCardData),
            });
            break;
          }
        }
      } catch (err) {
        console.error("Failed to load users:", err);
        dispatch({ type: "LOAD_ERROR" });
      }
    };

    loadUsers();
  }, [selectedRole]);

  // ─── PDF / Print helpers ──────────────────────────────────────────────

  const getCardHtml = (): string => {
    const cardEl = printRef.current?.querySelector(".id-card");
    return cardEl ? cardEl.outerHTML : "";
  };

  const getPrintStyles = (isBatch: boolean = false): string => `
    <style>
      @page {
        margin: ${isBatch ? "5mm" : "0"};
        size: ${isBatch ? "A4 landscape" : "120mm 180mm"};
      }
      @media print {
        html, body {
          margin: 0;
          padding: 0;
          display: flex;
          justify-content: center;
          align-items: flex-start;
          background: white;
          width: 100%;
          height: auto;
          overflow: visible;
          padding-left: 15mm;
        }
        body {
          -webkit-print-color-adjust: exact !important;
          print-color-adjust: exact !important;
        }
        .id-card {
          margin: ${isBatch ? "4px" : "10px auto"};
          overflow: visible !important;
          page-break-inside: avoid;
          box-shadow: none !important;
          flex-shrink: 0;
        }
        ${
          isBatch
            ? `
        .batch-container {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: 12px;
          padding: 10mm;
          width: 100%;
          justify-items: center;
        }
        `
            : ""
        }
      }
      * {
        box-sizing: border-box;
        -webkit-print-color-adjust: exact !important;
        print-color-adjust: exact !important;
      }
      body {
        margin: 0;
        padding: 0;
        font-family: system-ui, -apple-system, sans-serif;
        background: white;
      }
      .id-card {
        width: 300px;
        background: white;
        border-radius: 12px;
        border: 2px solid #e5e7eb;
        overflow: hidden;
      }
      .id-card * {
        box-sizing: border-box;
      }
      .break-words {
        overflow-wrap: break-word;
        word-wrap: break-word;
        word-break: break-word;
      }
    </style>
  `;

  const handlePrint = () => {
    const styleTags = Array.from(
      document.head.querySelectorAll(
        'link[rel="stylesheet"], link[rel="preconnect"], style',
      ),
    )
      .map((el) => el.outerHTML)
      .join("\n");

    if (batchMode === "batch") {
      if (selectedUsers.length === 0) return;
      const cardsToPrint = selectedUsers.slice(0, 4); // Limit to 4 cards per page
      const printWindow = window.open("", "_blank");
      if (!printWindow) {
        window.print();
        return;
      }

      const cardsHtml = cardsToPrint
        .map((card) => {
          const cardEl = document.getElementById(
            `card-${card.role}-${card.id}`,
          );
          return cardEl ? cardEl.outerHTML : "";
        })
        .join("");

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
          <head>
            <meta charset="utf-8">
            <meta name="viewport" content="width=210mm">
            <title>ID Cards - Batch</title>
            ${styleTags}
            ${getPrintStyles(true)}
          </head>
          <body>
            <div class="batch-container">
              ${cardsHtml}
            </div>
          </body>
        </html>
      `);
      printWindow.document.close();
      printWindow.focus();
      setTimeout(() => {
        printWindow.print();
        printWindow.close();
      }, 500);
    } else {
      const cardHtml = getCardHtml();
      if (!cardHtml) return;

      const printWindow = window.open("", "_blank");
      if (!printWindow) {
        window.print();
        return;
      }

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
          <head>
            <meta charset="utf-8">
            <meta name="viewport" content="width=210mm">
            <title>ID Card - ${selectedUser?.name ?? ""}</title>
            ${styleTags}
            ${getPrintStyles(false)}
          </head>
          <body>
            <div style="display:flex;justify-content:center;align-items:center;min-height:100vh;padding:20px;">
              ${cardHtml}
            </div>
          </body>
        </html>
      `);
      printWindow.document.close();
      printWindow.focus();
      setTimeout(() => {
        printWindow.print();
        printWindow.close();
      }, 500);
    }
  };

  const handlePdf = () => {
    const styleTags = Array.from(
      document.head.querySelectorAll(
        'link[rel="stylesheet"], link[rel="preconnect"], style',
      ),
    )
      .map((el) => el.outerHTML)
      .join("\n");

    if (batchMode === "batch") {
      if (selectedUsers.length === 0) return;
      const cardsToPrint = selectedUsers.slice(0, 4); // Limit to 4 cards per page
      const printWindow = window.open("", "_blank");
      if (!printWindow) return;

      const cardsHtml = cardsToPrint
        .map((card) => {
          const cardEl = document.getElementById(
            `card-${card.role}-${card.id}`,
          );
          return cardEl ? cardEl.outerHTML : "";
        })
        .join("");

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
          <head>
            <meta charset="utf-8">
            <meta name="viewport" content="width=210mm">
            <title>ID Cards - Batch</title>
            ${styleTags}
            ${getPrintStyles(true)}
          </head>
          <body>
            <div class="batch-container">
              ${cardsHtml}
            </div>
          </body>
        </html>
      `);
      printWindow.document.close();
      printWindow.focus();
      setTimeout(() => {
        printWindow.print();
        printWindow.close();
      }, 500);
    } else {
      const cardHtml = getCardHtml();
      if (!cardHtml) return;

      const printWindow = window.open("", "_blank");
      if (!printWindow) return;

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
          <head>
            <meta charset="utf-8">
            <meta name="viewport" content="width=210mm">
            <title>ID Card - ${selectedUser?.name ?? ""}</title>
            ${styleTags}
            ${getPrintStyles(false)}
          </head>
          <body>
            <div style="display:flex;justify-content:center;align-items:center;min-height:100vh;padding:20px;">
              ${cardHtml}
            </div>
          </body>
        </html>
      `);
      printWindow.document.close();
      printWindow.focus();
      setTimeout(() => {
        printWindow.print();
        printWindow.close();
      }, 500);
    }
  };

  const toggleUserSelection = (user: CardData) => {
    if (batchMode === "batch") {
      setSelectedUsers((prev) => {
        const exists = prev.find(
          (u) => u.id === user.id && u.role === user.role,
        );
        if (exists) {
          return prev.filter(
            (u) => !(u.id === user.id && u.role === user.role),
          );
        }
        if (prev.length >= 4) {
          return prev; // Limit to 4 cards
        }
        return [...prev, user];
      });
    }
  };

  // ─── Render ──────────────────────────────────────────────────────────

  if (!authUser || !role) {
    return (
      <div className="p-6 text-center text-muted">
        Unable to load user data.
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <PageHeader
        title="ID Card Generator"
        subtitle="Generate identification cards for existing users"
        action={
          <div className="flex gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={handlePrint}
              disabled={
                batchMode === "single"
                  ? !selectedUser
                  : selectedUsers.length === 0
              }
            >
              <Printer className="w-4 h-4 mr-1" /> Print
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={handlePdf}
              disabled={
                batchMode === "single"
                  ? !selectedUser
                  : selectedUsers.length === 0
              }
            >
              <Download className="w-4 h-4 mr-1" /> PDF
            </Button>
          </div>
        }
      />

      {/* Role Selector + Search */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-5 space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          {/* Role Selector */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium uppercase tracking-wider text-gray-500">
              Select Role
            </label>
            <select
              className="w-full px-3 py-2.5 rounded-lg text-sm bg-white border border-gray-300 text-gray-900 focus:outline-none focus:ring-2 focus:ring-accent/40 focus:border-accent"
              value={selectedRole}
              onChange={(e) => {
                const newRole = e.target.value as IdCardRole | "";
                setSelectedRole(newRole);
                setSelectedUsers([]);
                if (!newRole) {
                  dispatch({ type: "RESET" });
                } else {
                  dispatch({ type: "START_LOADING" });
                }
              }}
            >
              <option value="">— Choose a role —</option>
              <option value="Admin">Admin</option>
              <option value="Teacher">Teacher</option>
              <option value="Parent">Parent</option>
              <option value="Student">Student</option>
            </select>
          </div>

          {/* Batch Mode Toggle */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium uppercase tracking-wider text-gray-500">
              Generation Mode
            </label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => {
                  setBatchMode("single");
                  setSelectedUsers([]);
                }}
                className={`flex-1 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  batchMode === "single"
                    ? "bg-accent text-white"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                Single
              </button>
              <button
                type="button"
                onClick={() => {
                  setBatchMode("batch");
                  dispatch({ type: "SELECT_USER", user: null });
                }}
                className={`flex-1 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  batchMode === "batch"
                    ? "bg-accent text-white"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                Batch (4)
              </button>
            </div>
          </div>

          {/* Search */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium uppercase tracking-wider text-gray-500">
              {batchMode === "batch" ? "Select Users (max 4)" : "Select User"}
            </label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder={
                  selectedRole
                    ? `Search ${selectedRole}s by name, email or ID...`
                    : "Select a role first"
                }
                value={searchQuery}
                onChange={(e) =>
                  dispatch({ type: "SET_SEARCH", query: e.target.value })
                }
                disabled={!selectedRole}
                className="w-full pl-9 pr-3 py-2.5 rounded-lg text-sm bg-white border border-gray-300 text-gray-900 placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-accent/40 focus:border-accent disabled:opacity-50 disabled:cursor-not-allowed"
              />
              {searchQuery && (
                <button
                  onClick={() => dispatch({ type: "SET_SEARCH", query: "" })}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>
        </div>

        {/* User Results / Selected User */}
        {loadingUsers && (
          <div className="text-center py-4 text-sm text-gray-500">
            Loading users...
          </div>
        )}

        {!loadingUsers && selectedRole && filteredUsers.length === 0 && (
          <div className="text-center py-4 text-sm text-gray-400">
            {searchQuery
              ? "No matching users found."
              : `No ${selectedRole.toLowerCase()} users available.`}
          </div>
        )}

        {!loadingUsers && filteredUsers.length > 0 && (
          <div className="max-h-48 overflow-y-auto space-y-1 border border-gray-200 rounded-lg p-1">
            {filteredUsers.map((u) => {
              const isSelected =
                batchMode === "batch"
                  ? selectedUsers.some(
                      (su) => su.id === u.id && su.role === u.role,
                    )
                  : selectedUser?.id === u.id && selectedUser?.role === u.role;
              return (
                <button
                  key={`${u.role}-${u.id}`}
                  onClick={() => {
                    if (batchMode === "batch") {
                      toggleUserSelection(u);
                    } else {
                      dispatch({ type: "SELECT_USER", user: u });
                      dispatch({ type: "SET_SEARCH", query: u.name });
                    }
                  }}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors flex items-center gap-3 ${
                    isSelected
                      ? "bg-accent/10 text-accent border border-accent/30"
                      : "hover:bg-gray-50 text-gray-700 border border-transparent"
                  }`}
                >
                  {batchMode === "batch" && (
                    <div className="w-5 h-5 rounded border border-gray-300 flex items-center justify-center flex-shrink-0">
                      {isSelected && (
                        <div className="w-3 h-3 rounded-sm bg-accent" />
                      )}
                    </div>
                  )}
                  <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center text-xs font-bold text-gray-600 flex-shrink-0 overflow-hidden">
                    {u.avatar_url ? (
                      <img
                        src={profileUrl(u.avatar_url)!}
                        alt=""
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      u.name[0]?.toUpperCase()
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-medium truncate">{u.name}</div>
                    <div className="text-xs text-gray-400 truncate">
                      {u.code && <span className="font-mono">{u.code} • </span>}
                      {u.email}
                    </div>
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </div>

      {/* ID Card Preview */}
      {batchMode === "batch" ? (
        selectedUsers.length > 0 ? (
          <div
            ref={printRef}
            className="grid grid-cols-1 sm:grid-cols-2 gap-6 justify-items-center py-4"
          >
            {selectedUsers.map((user) => (
              <IdCardRenderer key={`${user.role}-${user.id}`} card={user} />
            ))}
          </div>
        ) : (
          <div className="flex justify-center py-12">
            <div className="text-center text-gray-400">
              <div className="w-24 h-24 rounded-full bg-gray-100 mx-auto mb-4 flex items-center justify-center">
                <Search className="w-8 h-8 text-gray-300" />
              </div>
              <p className="text-sm">
                Select up to 4 users to generate ID cards in batch mode
              </p>
            </div>
          </div>
        )
      ) : selectedUser ? (
        <div ref={printRef} className="flex justify-center py-4">
          {role === "Admin" ? (
            <AdminPhotoOverlay
              card={selectedUser}
              onUploaded={(url) =>
                dispatch({
                  type: "UPDATE_USER_AVATAR",
                  userId: selectedUser.id,
                  role: selectedUser.role,
                  avatarUrl: url,
                })
              }
            />
          ) : (
            <IdCardRenderer card={selectedUser} />
          )}
        </div>
      ) : (
        <div className="flex justify-center py-12">
          <div className="text-center text-gray-400">
            <div className="w-24 h-24 rounded-full bg-gray-100 mx-auto mb-4 flex items-center justify-center">
              <Search className="w-8 h-8 text-gray-300" />
            </div>
            <p className="text-sm">
              Select a role and user to generate an ID card
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
