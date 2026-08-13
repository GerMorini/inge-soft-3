import { useCallback, useState } from 'react'
import { LoginForm } from './auth/LoginForm'
import { RegisterForm } from './auth/RegisterForm'
import { SessionStatus } from './auth/SessionStatus'
import { readAccessToken } from './auth/session'

export default function App() {
  const [view, setView] = useState<'register' | 'login'>('register')
  const [authenticated, setAuthenticated] = useState(() => readAccessToken() !== null)
  const unauthenticate = useCallback(() => setAuthenticated(false), [])

  return (
    <main className="min-h-screen bg-base-200 px-4 py-10">
      <section className="card mx-auto max-w-3xl bg-base-100 shadow-xl" aria-labelledby="page-title">
        <div className="card-body pb-0">
          <p className="text-sm font-semibold text-primary">Aplicación académica</p>
          <h1 className="card-title text-3xl" id="page-title">{authenticated ? 'Tu sesión' : view === 'register' ? 'Crear cuenta' : 'Iniciar sesión'}</h1>
          {!authenticated && (
            <div className="join" aria-label="Elegir formulario">
              <button className={`btn join-item ${view === 'register' ? 'btn-active' : ''}`} onClick={() => setView('register')} type="button">Registro</button>
              <button className={`btn join-item ${view === 'login' ? 'btn-active' : ''}`} onClick={() => setView('login')} type="button">Login</button>
            </div>
          )}
        </div>
        {authenticated ? (
          <SessionStatus onUnauthenticated={unauthenticate} />
        ) : view === 'register' ? (
          <RegisterForm onRegistered={() => setView('login')} />
        ) : (
          <LoginForm onAuthenticated={() => setAuthenticated(true)} />
        )}
      </section>
    </main>
  )
}
