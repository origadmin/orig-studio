/**
 * Authentication type definitions for the AuthProvider + AuthContext system.
 *
 * These types define the contract between the AuthProvider (state source)
 * and its consumers (useAuth hook, Router Context, route guards).
 */

/** User information stored in auth state */
export interface User {
  id: number;
  username: string;
  displayName: string;
  avatarUrl?: string;
  roles: string[];
}

/** Read-only portion of the authentication state */
export interface AuthState {
  /** Current access token, null when not authenticated */
  token: string | null;
  /** Current user info, null when not authenticated */
  user: User | null;
  /** Whether the user is authenticated (token exists, not expired, user present) */
  isAuthenticated: boolean;
  /** Whether the user has the admin role */
  isAdmin: boolean;
  /** Whether a token refresh is currently in progress */
  isRefreshing: boolean;
  /** Whether the initial load from localStorage has completed */
  isInitialized: boolean;
}

/** Action methods exposed by the AuthProvider */
export interface AuthActions {
  /** Login: write token + user info to state and localStorage */
  login: (token: string, refreshToken: string, user: User) => void;
  /** Logout: clear state and localStorage */
  logout: () => void;
  /** Manually trigger a token refresh, returns success status */
  refreshAuth: () => Promise<boolean>;
}

/** Complete value type of the AuthContext */
export interface AuthContextValue extends AuthState, AuthActions {}

/** Props for the AuthProvider component */
export interface AuthProviderProps {
  children: React.ReactNode;
}
