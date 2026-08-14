import { useCallback, useRef, useState, type KeyboardEvent } from "react";
import {
  CalendarDays,
  CircleUserRound,
  Dumbbell,
  ListChecks,
  LogOut,
} from "lucide-react";
import { AuthShell } from "./auth/AuthShell";
import { LoginForm } from "./auth/LoginForm";
import { RegisterForm } from "./auth/RegisterForm";
import { SessionStatus } from "./auth/SessionStatus";
import { readAccessToken } from "./auth/session";
import { ExercisesView } from "./routines/ExercisesView";
import { RoutinesView } from "./routines/RoutinesView";
import { SessionsView } from "./routines/SessionsView";
import { ModalDialog } from "./routines/components/ModalDialog";

const workspaceViews = ["routines", "sessions", "exercises"] as const;
type WorkspaceView = (typeof workspaceViews)[number];
const labels = {
  routines: "Rutinas",
  sessions: "Sesiones",
  exercises: "Ejercicios",
};
const icons = {
  routines: CalendarDays,
  sessions: ListChecks,
  exercises: Dumbbell,
};

export default function App() {
  const [authView, setAuthView] = useState<"register" | "login">("login");
  const [workspaceView, setWorkspaceView] = useState<WorkspaceView>("routines");
  const [workspaceDirty, setWorkspaceDirty] = useState(false);
  const [pendingWorkspace, setPendingWorkspace] = useState<WorkspaceView>();
  const [authenticated, setAuthenticated] = useState(
    () => readAccessToken() !== null,
  );
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const unauthenticate = useCallback(() => {
    setWorkspaceDirty(false);
    setAuthenticated(false);
  }, []);
  const handleDirtyChange = useCallback(
    (dirty: boolean) => setWorkspaceDirty(dirty),
    [],
  );
  function changeWorkspace(next: WorkspaceView) {
    if (next === workspaceView) return true;
    if (workspaceDirty) {
      setPendingWorkspace(next);
      return false;
    }
    setWorkspaceDirty(false);
    setWorkspaceView(next);
    return true;
  }
  function confirmWorkspaceChange() {
    if (!pendingWorkspace) return;
    const next = pendingWorkspace;
    setPendingWorkspace(undefined);
    setWorkspaceDirty(false);
    setWorkspaceView(next);
    requestAnimationFrame(() =>
      tabRefs.current[workspaceViews.indexOf(next)]?.focus(),
    );
  }
  function handleKey(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    let next = index;
    if (event.key === "ArrowRight") next = (index + 1) % 3;
    else if (event.key === "ArrowLeft") next = (index + 2) % 3;
    else if (event.key === "Home") next = 0;
    else if (event.key === "End") next = 2;
    else return;
    event.preventDefault();
    if (changeWorkspace(workspaceViews[next])) tabRefs.current[next]?.focus();
  }

  if (!authenticated)
    return (
      <AuthShell
        title={authView === "login" ? "Iniciar sesión" : "Crear cuenta"}
      >
        {authView === "login" ? (
          <LoginForm
            onAuthenticated={() => setAuthenticated(true)}
            onShowRegister={() => setAuthView("register")}
          />
        ) : (
          <RegisterForm
            onRegistered={() => setAuthView("login")}
            onShowLogin={() => setAuthView("login")}
          />
        )}
      </AuthShell>
    );

  return (
    <SessionStatus onUnauthenticated={unauthenticate}>
      {({ currentUser, logout }) => (
        <>
          <a
            className="fixed left-4 top-4 z-50 -translate-y-[200%] rounded-lg bg-base-content px-4 py-3 text-base-100 focus:translate-y-0"
            href="#workspace-content"
          >
            Saltar al contenido
          </a>
          <nav
            className="sticky top-0 z-30 border-b border-base-300 bg-base-100/95 backdrop-blur"
            aria-label="Navegación principal"
          >
            <div className="mx-auto grid max-w-7xl grid-cols-1 items-center gap-3 px-4 py-3 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]">
              <div className="flex items-center gap-2 justify-self-start text-2xl font-black">
                <Dumbbell aria-hidden="true" className="text-primary" />
                FitPro
              </div>
              <div
                role="tablist"
                aria-label="Apartados"
                className="flex flex-wrap justify-center gap-1 justify-self-center"
              >
                {workspaceViews.map((view, index) => {
                  const Icon = icons[view];
                  return (
                    <button
                      key={view}
                      ref={(node) => {
                        tabRefs.current[index] = node;
                      }}
                      role="tab"
                      aria-selected={workspaceView === view}
                      aria-current={workspaceView === view ? "page" : undefined}
                      tabIndex={workspaceView === view ? 0 : -1}
                      className={`btn btn-ghost ${workspaceView === view ? "text-secondary" : ""}`}
                      type="button"
                      onClick={() => changeWorkspace(view)}
                      onKeyDown={(event) => handleKey(event, index)}
                    >
                      <Icon aria-hidden="true" size={18} />
                      {labels[view]}
                    </button>
                  );
                })}
              </div>
              <div
                role="group"
                className="flex min-w-0 items-center gap-2 justify-self-end"
                aria-label="Controles de sesión"
              >
                <CircleUserRound
                  aria-hidden="true"
                  className="shrink-0 text-primary"
                  size={20}
                />
                <strong className="max-w-36 truncate">
                  {currentUser.username}
                </strong>
                <button
                  className="btn btn-ghost btn-sm min-h-11"
                  onClick={logout}
                  type="button"
                  aria-label={`Cerrar sesión de ${currentUser.username}`}
                >
                  <LogOut aria-hidden="true" size={18} />
                  <span className="hidden lg:inline">Cerrar sesión</span>
                </button>
              </div>
            </div>
          </nav>
          <main
            id="workspace-content"
            className="mx-auto max-w-7xl px-4 py-8"
            tabIndex={-1}
          >
            {workspaceView === "routines" && (
              <RoutinesView
                onUnauthenticated={unauthenticate}
                onDirtyChange={handleDirtyChange}
              />
            )}
            {workspaceView === "sessions" && (
              <SessionsView
                onUnauthenticated={unauthenticate}
                onDirtyChange={handleDirtyChange}
              />
            )}
            {workspaceView === "exercises" && (
              <ExercisesView
                onUnauthenticated={unauthenticate}
                onDirtyChange={handleDirtyChange}
              />
            )}
          </main>
          <ModalDialog
            open={pendingWorkspace !== undefined}
            role="alertdialog"
            title="¿Salir del wizard?"
            description="Los cambios se perderán, ¿seguro deseas salir?"
            onRequestClose={() => setPendingWorkspace(undefined)}
          >
            <div className="flex flex-wrap justify-end gap-3">
              <button
                type="button"
                className="btn btn-ghost"
                onClick={() => setPendingWorkspace(undefined)}
              >
                Seguir editando
              </button>
              <button
                type="button"
                className="btn btn-error"
                onClick={confirmWorkspaceChange}
              >
                Salir sin guardar
              </button>
            </div>
          </ModalDialog>
        </>
      )}
    </SessionStatus>
  );
}
