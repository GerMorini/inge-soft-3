import {
  ApiError,
  type CurrentUser,
  type ErrorBody,
  type LoginInput,
  type LoginResult,
  type RegisterInput,
  type RegisteredUser,
} from './types'
import { clearAccessToken, readAccessToken } from './session'

export async function register(input: RegisterInput): Promise<RegisteredUser> {
  const response = await fetch('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })

  if (!response.ok) {
    const payload = (await response.json()) as { error: ErrorBody }
    throw new ApiError(response.status, payload.error)
  }
  return (await response.json()) as RegisteredUser
}

export async function login(input: LoginInput): Promise<LoginResult> {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!response.ok) {
    const payload = (await response.json()) as { error: ErrorBody }
    throw new ApiError(response.status, payload.error)
  }
  return (await response.json()) as LoginResult
}

export async function getCurrentUser(): Promise<CurrentUser> {
  const token = readAccessToken()
  if (!token) {
    throw new ApiError(401, { code: 'invalid_token', message: 'No hay una sesión activa.' })
  }
  const response = await fetch('/api/auth/me', {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!response.ok) {
    const payload = (await response.json()) as { error: ErrorBody }
    if (response.status === 401) clearAccessToken()
    throw new ApiError(response.status, payload.error)
  }
  return (await response.json()) as CurrentUser
}
