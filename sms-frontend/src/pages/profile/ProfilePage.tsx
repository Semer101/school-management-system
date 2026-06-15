import { useRef, useState, useEffect, type FormEvent } from "react";
import { Phone, Camera } from "lucide-react";
import { useAuth } from "../../hooks/useAuth";
import { changePassword, uploadAvatar } from "../../api/me";
import { Input } from "../../components/ui/Input";
import { Button } from "../../components/ui/Button";
import { Badge, roleBadgeVariant } from "../../components/ui/Badge";
import { PageHeader } from "../../components/ui/PageHeader";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";
const MAX_AVATAR_BYTES = 2 * 1024 * 1024; // 2 MB — must match backend MaxAvatarSize

export default function ProfilePage() {
  const { user, refreshUser } = useAuth();
  const fileRef = useRef<HTMLInputElement>(null);
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [success, setSuccess] = useState("");
  const [error, setError] = useState("");
  const [localPreview, setLocalPreview] = useState<string | null>(null);

  // Derive avatar to show: local optimistic preview > stored url > null
  const avatarSrc =
    localPreview ?? (user?.avatar_url ? `${API_BASE}${user.avatar_url}` : null);

  // Revoke the object URL when component unmounts or preview changes
  useEffect(() => {
    return () => {
      if (localPreview) URL.revokeObjectURL(localPreview);
    };
  }, [localPreview]);

  const handleAvatar = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Client-side size guard (mirrors backend 2 MB limit)
    if (file.size > MAX_AVATAR_BYTES) {
      setError("Image is too large. Maximum size is 2 MB.");
      if (fileRef.current) fileRef.current.value = "";
      return;
    }

    // Show optimistic preview immediately
    const preview = URL.createObjectURL(file);
    setLocalPreview(preview);

    setUploading(true);
    setError("");
    setSuccess("");
    try {
      await uploadAvatar(file);
      await refreshUser(); // sync the Redux store with new avatar_url
      setLocalPreview(null); // let the store value take over
      setSuccess("Profile picture updated!");
    } catch (err: unknown) {
      setLocalPreview(null); // revert preview on failure
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data
          ?.error ?? "Failed to upload image.";
      setError(msg);
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = "";
    }
  };

  const handlePassword = async (e: FormEvent) => {
    e.preventDefault();
    if (next !== confirm) {
      setError("Passwords do not match.");
      return;
    }
    setError("");
    setSuccess("");
    setLoading(true);
    try {
      await changePassword(current, next);
      setSuccess("Password updated successfully.");
      setCurrent("");
      setNext("");
      setConfirm("");
    } catch (err: unknown) {
      setError(
        (err as { response?: { data?: { error?: string } } })?.response?.data
          ?.error ?? "Failed to update password.",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto">
      <PageHeader title="My Profile" />

      <div className="bg-surface border border-surface-border rounded-2xl p-6 mb-6">
        <div className="flex items-center gap-4 mb-4">
          <div className="relative">
            {avatarSrc ? (
              <img
                src={avatarSrc}
                alt=""
                className="w-16 h-16 rounded-full object-cover border-2 border-accent/30"
              />
            ) : (
              <div className="w-16 h-16 rounded-full bg-accent/20 flex items-center justify-center text-accent text-2xl font-bold">
                {user?.name?.[0]?.toUpperCase()}
              </div>
            )}
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              className="absolute -bottom-1 -right-1 p-1.5 rounded-full bg-accent text-white shadow disabled:opacity-50"
              disabled={uploading}
              title="Change profile picture"
            >
              <Camera className="w-3 h-3" />
            </button>
            <input
              ref={fileRef}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              className="hidden"
              onChange={handleAvatar}
            />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-foreground">
              {user?.name}
            </h2>
            <p className="text-sm text-muted">{user?.email}</p>
            {user?.role && (
              <div className="mt-1">
                <Badge
                  label={user.role}
                  variant={roleBadgeVariant(user.role)}
                />
              </div>
            )}
          </div>
        </div>
        {uploading && <p className="text-xs text-muted mb-2">Uploading…</p>}
        <p className="text-[11px] text-muted/70 mt-1">
          Photo: JPG, PNG or WebP · max&nbsp;2&nbsp;MB
        </p>
        {user?.phone && (
          <p className="text-sm text-muted flex items-center gap-2">
            <Phone className="w-4 h-4 text-accent shrink-0" />
            {user.phone}
          </p>
        )}
        <p className="text-xs text-muted mt-2">
          Member since{" "}
          {user?.created_at
            ? new Date(user.created_at).toLocaleDateString()
            : "—"}
        </p>
      </div>

      <div className="bg-surface border border-surface-border rounded-2xl p-6">
        <h2 className="text-base font-semibold text-foreground mb-4">
          Change Password
        </h2>
        <form onSubmit={handlePassword} className="flex flex-col gap-4">
          <Input
            label="Current Password"
            type="password"
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            required
          />
          <Input
            label="New Password"
            type="password"
            value={next}
            onChange={(e) => setNext(e.target.value)}
            required
          />
          <Input
            label="Confirm New Password"
            type="password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            required
          />
          {error && <p className="text-sm text-danger">{error}</p>}
          {success && <p className="text-sm text-success">{success}</p>}
          <Button type="submit" loading={loading}>
            Update Password
          </Button>
        </form>
      </div>
    </div>
  );
}
