import { ApiError, type ErrorBody } from '../auth/types'
import { clearAccessToken, readAccessToken } from '../auth/session'
import type {
  Exercise,
  ExerciseInput,
  RoutineDetail,
  RoutineInput,
  RoutineSummary,
  SessionDetail,
  SessionInput,
  SessionSummary,
} from './types'

export type UnauthorizedHandler = () => void

async function authorizedRequest<T>(
  path: string,
  options: RequestInit = {},
  onUnauthorized?: UnauthorizedHandler,
): Promise<T> {
  const token = readAccessToken()
  if (!token) {
    onUnauthorized?.()
    throw new ApiError(401, { code: 'invalid_token', message: 'No hay una sesión activa.' })
  }
  const headers = new Headers(options.headers)
  headers.set('Authorization', `Bearer ${token}`)
  if (options.body) headers.set('Content-Type', 'application/json')
  const response = await fetch(path, { ...options, headers })
  if (response.status === 401) {
    clearAccessToken()
    onUnauthorized?.()
  }
  if (!response.ok) {
    let error: ErrorBody = { code: 'internal_error', message: 'No se pudo completar la operación.' }
    try {
      error = ((await response.json()) as { error: ErrorBody }).error
    } catch {
      // Mantiene mensaje seguro cuando respuesta técnica no contiene JSON.
    }
    throw new ApiError(response.status, error)
  }
  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

export const listExercises = (onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<Exercise[]>('/api/exercises', {}, onUnauthorized)

export const getExercise = (id: number, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<Exercise>(`/api/exercises/${id}`, {}, onUnauthorized)

export const createExercise = (input: ExerciseInput, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<Exercise>('/api/exercises', { method: 'POST', body: JSON.stringify(input) }, onUnauthorized)

export const updateExercise = (id: number, input: ExerciseInput, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<Exercise>(`/api/exercises/${id}`, { method: 'PUT', body: JSON.stringify(input) }, onUnauthorized)

export const deleteExercise = (id: number, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<void>(`/api/exercises/${id}`, { method: 'DELETE' }, onUnauthorized)

export const listSessions = (onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<SessionSummary[]>('/api/sessions', {}, onUnauthorized)

export const getSession = (id: number, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<SessionDetail>(`/api/sessions/${id}`, {}, onUnauthorized)

export const createSession = (input: SessionInput, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<SessionDetail>('/api/sessions', { method: 'POST', body: JSON.stringify(input) }, onUnauthorized)

export const updateSession = (id: number, input: SessionInput, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<SessionDetail>(`/api/sessions/${id}`, { method: 'PUT', body: JSON.stringify(input) }, onUnauthorized)

export const deleteSession = (id: number, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<void>(`/api/sessions/${id}`, { method: 'DELETE' }, onUnauthorized)

export const listRoutines = (onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<RoutineSummary[]>('/api/routines', {}, onUnauthorized)

export const getRoutine = (id: number, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<RoutineDetail>(`/api/routines/${id}`, {}, onUnauthorized)

export const createRoutine = (input: RoutineInput, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<RoutineDetail>('/api/routines', { method: 'POST', body: JSON.stringify(input) }, onUnauthorized)

export const updateRoutine = (id: number, input: RoutineInput, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<RoutineDetail>(`/api/routines/${id}`, { method: 'PUT', body: JSON.stringify(input) }, onUnauthorized)

export const deleteRoutine = (id: number, onUnauthorized?: UnauthorizedHandler) =>
  authorizedRequest<void>(`/api/routines/${id}`, { method: 'DELETE' }, onUnauthorized)
