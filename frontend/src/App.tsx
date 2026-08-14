import { useCallback, useRef, useState, type KeyboardEvent } from 'react'
import { LoginForm } from './auth/LoginForm'
import { RegisterForm } from './auth/RegisterForm'
import { SessionStatus } from './auth/SessionStatus'
import { readAccessToken } from './auth/session'
import { ExercisesView } from './routines/ExercisesView'
import { RoutinesView } from './routines/RoutinesView'
import { SessionsView } from './routines/SessionsView'

const workspaceViews = ['routines', 'sessions', 'exercises'] as const
type WorkspaceView = typeof workspaceViews[number]

export default function App() {
  const [authView, setAuthView] = useState<'register' | 'login'>('register')
  const [workspaceView, setWorkspaceView] = useState<WorkspaceView>('routines')
  const [workspaceDirty, setWorkspaceDirty] = useState(false)
  const [authenticated, setAuthenticated] = useState(() => readAccessToken() !== null)
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([])
  const unauthenticate = useCallback(() => {
    setWorkspaceDirty(false)
    setAuthenticated(false)
  }, [])
  const handleDirtyChange = useCallback((dirty: boolean) => setWorkspaceDirty(dirty), [])

  function changeWorkspace(nextView: WorkspaceView) {
    if (nextView === workspaceView) return true
    if (workspaceDirty && !window.confirm('Tenés cambios sin guardar. ¿Querés descartarlos?')) return false
    setWorkspaceDirty(false)
    setWorkspaceView(nextView)
    return true
  }

  // Flechas recorren pestañas sin añadir paradas redundantes al orden Tab.
  // WCAG 2.2 - 2.1.1 Keyboard, 4.1.2 Name, Role, Value.
  function handleTabKey(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    let nextIndex = index
    if (event.key === 'ArrowRight') nextIndex = (index + 1) % workspaceViews.length
    else if (event.key === 'ArrowLeft') nextIndex = (index - 1 + workspaceViews.length) % workspaceViews.length
    else if (event.key === 'Home') nextIndex = 0
    else if (event.key === 'End') nextIndex = workspaceViews.length - 1
    else return
    event.preventDefault()
    if (changeWorkspace(workspaceViews[nextIndex])) tabRefs.current[nextIndex]?.focus()
  }

  return (
    <main className="min-h-screen bg-base-200 px-4 py-10">
      <a className="fixed left-4 top-4 z-50 -translate-y-[200%] rounded-lg bg-base-content px-4 py-3 text-base-100 focus:translate-y-0" href="#workspace-content">Saltar al contenido</a>
      <section className="card mx-auto max-w-6xl bg-base-100 shadow-xl" aria-labelledby="page-title">
        <div className="card-body pb-0">
          <p className="text-sm font-semibold text-primary">Aplicación académica</p>
          <h1 className="card-title text-3xl" id="page-title">{authenticated ? 'Entrenamiento personal' : authView === 'register' ? 'Crear cuenta' : 'Iniciar sesión'}</h1>
          {!authenticated && (
            <div className="join" aria-label="Elegir formulario">
              <button className={`btn join-item ${authView === 'register' ? 'btn-active' : ''}`} onClick={() => setAuthView('register')} type="button">Registro</button>
              <button className={`btn join-item ${authView === 'login' ? 'btn-active' : ''}`} onClick={() => setAuthView('login')} type="button">Login</button>
            </div>
          )}
        </div>
        {authenticated ? (
          <SessionStatus onUnauthenticated={unauthenticate}>
            <nav aria-label="Contenido de entrenamiento">
              <div className="tabs tabs-box" role="tablist" aria-label="Rutinas, sesiones y ejercicios">
                {workspaceViews.map((item, index) => {
                  const label = item === 'routines' ? 'Rutinas' : item === 'sessions' ? 'Sesiones' : 'Ejercicios'
                  return (
                    <button
                      aria-controls={`panel-${item}`}
                      aria-selected={workspaceView === item}
                      className={`tab ${workspaceView === item ? 'tab-active' : ''}`}
                      id={`tab-${item}`}
                      key={item}
                      onClick={(event) => { if (!changeWorkspace(item)) event.preventDefault() }}
                      onKeyDown={(event) => handleTabKey(event, index)}
                      ref={(element) => { tabRefs.current[index] = element }}
                      role="tab"
                      tabIndex={workspaceView === item ? 0 : -1}
                      type="button"
                    >{label}</button>
                  )
                })}
              </div>
            </nav>
            <div id="workspace-content">
              <div aria-labelledby={`tab-${workspaceView}`} id={`panel-${workspaceView}`} role="tabpanel" tabIndex={0}>
                {workspaceView === 'routines' && <RoutinesView onUnauthenticated={unauthenticate} onDirtyChange={handleDirtyChange} />}
                {workspaceView === 'sessions' && <SessionsView onUnauthenticated={unauthenticate} onDirtyChange={handleDirtyChange} />}
                {workspaceView === 'exercises' && <ExercisesView onUnauthenticated={unauthenticate} onDirtyChange={handleDirtyChange} />}
              </div>
            </div>
          </SessionStatus>
        ) : authView === 'register' ? (
          <RegisterForm onRegistered={() => setAuthView('login')} />
        ) : (
          <LoginForm onAuthenticated={() => setAuthenticated(true)} />
        )}
      </section>
    </main>
  )
}
