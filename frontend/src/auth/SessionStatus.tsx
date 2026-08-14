import { useEffect, useState, type ReactNode } from "react";
import { getCurrentUser } from "./api";
import {
  clearAccessToken,
  isAccessTokenExpired,
  readAccessToken,
} from "./session";
import type { CurrentUser } from "./types";

interface SessionStatusProps {
  onUnauthenticated: () => void;
  children?:
    | ReactNode
    | ((session: {
        currentUser: CurrentUser;
        logout: () => void;
      }) => ReactNode);
}

export function SessionStatus({
  onUnauthenticated,
  children,
}: SessionStatusProps) {
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = readAccessToken();
    if (!token || isAccessTokenExpired(token)) {
      clearAccessToken();
      setLoading(false);
      onUnauthenticated();
      return;
    }

    void getCurrentUser()
      .then(setCurrentUser)
      .catch(() => onUnauthenticated())
      .finally(() => setLoading(false));
  }, [onUnauthenticated]);

  function logout() {
    clearAccessToken();
    onUnauthenticated();
  }

  if (loading) {
    return (
      <div
        className="flex min-h-screen items-center justify-center"
        role="status"
      >
        <span className="loading loading-spinner" aria-hidden="true" />
        <span>Validando sesión</span>
      </div>
    );
  }
  if (!currentUser) return null;

  return (
    <div className="min-h-screen bg-base-200">
      {typeof children === "function"
        ? children({ currentUser, logout })
        : children}
    </div>
  );
}
