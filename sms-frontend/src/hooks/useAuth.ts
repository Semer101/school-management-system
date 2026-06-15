import { useAppDispatch, useAppSelector } from "../store/hooks";
import {
  login as loginThunk,
  logout as logoutThunk,
  clearError,
  refreshUser as refreshUserThunk,
} from "../store/authSlice";

export function useAuth() {
  const dispatch = useAppDispatch();
  const { user, loading, initialized, error } = useAppSelector((s) => s.auth);

  return {
    user,
    loading,
    initialized,
    error,
    login: (email: string, password: string) =>
      dispatch(loginThunk({ email, password })).unwrap(),
    logout: () => dispatch(logoutThunk()),
    clearError: () => dispatch(clearError()),
    refreshUser: () => dispatch(refreshUserThunk()),
  };
}
