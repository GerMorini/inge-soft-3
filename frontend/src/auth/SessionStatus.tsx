import { useEffect, useState } from 'react'
import { getCurrentUser } from './api'
import { clearAccessToken, isAccessTokenExpired, readAccessToken } from './session'
import type { CurrentUser } from './types'

interface SessionStatusProps {
  onUnauthenticated: () => void
}

export function SessionStatus({ onUnauthenticated }: SessionStatusProps) {
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = readAccessToken()
    if (!token || isAccessTokenExpired(token)) {
      clearAccessToken()
      setLoading(false)
      onUnauthenticated()
      return
    }

    void getCurrentUser()
      .then(setCurrentUser)
      .catch(() => onUnauthenticated())
      .finally(() => setLoading(false))
  }, [onUnauthenticated])

  function logout() {
    clearAccessToken()
    onUnauthenticated()
  }

  if (loading) {
    return (
      <div className="card-body">
        <span className="loading loading-spinner" aria-label="Validando sesión" />
      </div>
    )
  }
  if (!currentUser) return null

  return (
    <div className="card-body gap-4">
      <div className="alert alert-success" role="status">
        Sesión activa: <strong>{currentUser.username}</strong>
      </div>
      <button className="btn btn-outline" onClick={logout} type="button">Cerrar sesión</button>
    </div>
  )
}
