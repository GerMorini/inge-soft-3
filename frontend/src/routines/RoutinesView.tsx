import { useEffect, useState, type FormEvent } from 'react'
import { ApiError } from '../auth/types'
import { createRoutine, deleteRoutine, getRoutine, listRoutines, listSessions } from './api'
import type { FieldErrors, RoutineDetail, RoutineSummary, SessionSummary } from './types'
import { SessionDetailCard } from './SessionsView'

interface RoutinesViewProps {
  onUnauthenticated: () => void
}

const weekdays = ['', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo']

interface Assignment {
  session: SessionSummary
  day: number
}

export function RoutinesView({ onUnauthenticated }: RoutinesViewProps) {
  const [routines, setRoutines] = useState<RoutineSummary[]>([])
  const [sessions, setSessions] = useState<SessionSummary[]>([])
  const [selected, setSelected] = useState<RoutineDetail | null>(null)
  const [assignments, setAssignments] = useState<Assignment[]>([])
  const [sessionID, setSessionID] = useState('')
  const [day, setDay] = useState('1')
  const [form, setForm] = useState({ name: '', description: '' })
  const [fields, setFields] = useState<FieldErrors>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [routineItems, sessionItems] = await Promise.all([
        listRoutines(onUnauthenticated),
        listSessions(onUnauthenticated),
      ])
      setRoutines(routineItems)
      setSessions(sessionItems)
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudieron cargar las rutinas.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  function addAssignment() {
    const session = sessions.find((item) => item.id === Number(sessionID))
    const selectedDay = Number(day)
    if (!session) return
    if (assignments.some((item) => item.session.id === session.id && item.day === selectedDay)) {
      setError(`${session.name} ya está asignada al ${weekdays[selectedDay].toLowerCase()}.`)
      return
    }
    setAssignments((current) => [...current, { session, day: selectedDay }])
    setSessionID('')
    setError('')
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    const nextFields: FieldErrors = {}
    if (!form.name) nextFields.name = ['Este campo es obligatorio.']
    else if (form.name.trim() !== form.name || /\s{2,}|[\t\n\r]/.test(form.name) || form.name.length > 100) nextFields.name = ['Usá hasta 100 caracteres y un espacio entre palabras.']
    if (form.description && (form.description.trim() !== form.description || /\s{2,}|[\t\n\r]/.test(form.description) || form.description.length > 500)) nextFields.description = ['Usá hasta 500 caracteres y un espacio entre palabras.']
    if (Object.keys(nextFields).length > 0) {
      setFields(nextFields)
      setError('Revisá los campos indicados.')
      return
    }
    setFields({})
    setError('')
    try {
      const created = await createRoutine({
        ...form,
        sessions: assignments.map((item) => ({ sessionId: item.session.id, day: item.day })),
      }, onUnauthenticated)
      setRoutines((current) => [...current, created].sort((a, b) => a.id - b.id))
      setSelected(created)
      setAssignments([])
      setForm({ name: '', description: '' })
      setMessage(`Rutina ${created.name} creada.`)
    } catch (reason) {
      if (reason instanceof ApiError) setFields(reason.body.fields ?? {})
      if (!(reason instanceof ApiError && reason.status === 401)) setError(reason instanceof Error ? reason.message : 'No se pudo crear la rutina.')
    }
  }

  async function inspect(id: number) {
    setError('')
    try {
      setSelected(await getRoutine(id, onUnauthenticated))
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudo cargar la rutina.')
    }
  }

  async function remove(routine: RoutineSummary) {
    if (!window.confirm(`¿Eliminar la rutina ${routine.name}?`)) return
    setError('')
    try {
      await deleteRoutine(routine.id, onUnauthenticated)
      setRoutines((current) => current.filter((item) => item.id !== routine.id))
      setSelected((current) => current?.id === routine.id ? null : current)
      setMessage(`Rutina ${routine.name} eliminada.`)
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudo eliminar la rutina.')
    }
  }

  return (
    <section aria-labelledby="routines-title" className="space-y-6">
      <h2 className="text-2xl font-bold" id="routines-title">Rutinas</h2>
      {message && <p className="alert alert-success" role="status">{message}</p>}
      {error && <p className="alert alert-error" role="alert" tabIndex={-1}>{error}</p>}

      <form aria-label="Crear rutina" className="card bg-base-200" onSubmit={submit} noValidate>
        <div className="card-body gap-4">
          <h3 className="card-title">Crear rutina</h3>
          <label className="fieldset">
            <span className="fieldset-legend">Nombre *</span>
            <input className={`input w-full ${fields.name ? 'input-error' : ''}`} value={form.name} required aria-invalid={fields.name ? 'true' : undefined} aria-describedby={fields.name ? 'routine-name-error' : undefined} onChange={(event) => setForm({ ...form, name: event.target.value })} />
            {fields.name && <span className="text-error" id="routine-name-error">{fields.name.join(' ')}</span>}
          </label>
          <label className="fieldset">
            <span className="fieldset-legend">Descripción</span>
            <textarea className={`textarea w-full ${fields.description ? 'textarea-error' : ''}`} value={form.description} aria-invalid={fields.description ? 'true' : undefined} aria-describedby={fields.description ? 'routine-description-error' : undefined} onChange={(event) => setForm({ ...form, description: event.target.value })} />
            {fields.description && <span className="text-error" id="routine-description-error">{fields.description.join(' ')}</span>}
          </label>

          <fieldset className="fieldset rounded-box border border-base-300 p-4">
            <legend className="fieldset-legend px-2">Sesiones de la rutina</legend>
            <div className="grid gap-2 sm:grid-cols-[1fr_1fr_auto]">
              <label className="fieldset">
                <span className="fieldset-legend">Sesión</span>
                <select className="select w-full" value={sessionID} onChange={(event) => setSessionID(event.target.value)}>
                  <option value="">Elegí una sesión</option>
                  {sessions.map((session) => <option key={session.id} value={session.id}>{session.name}</option>)}
                </select>
              </label>
              <label className="fieldset">
                <span className="fieldset-legend">Día</span>
                <select className="select w-full" value={day} onChange={(event) => setDay(event.target.value)}>
                  {weekdays.slice(1).map((weekday, index) => <option key={weekday} value={index + 1}>{weekday}</option>)}
                </select>
              </label>
              <button className="btn self-end" type="button" disabled={!sessionID} onClick={addAssignment}>Agregar sesión</button>
            </div>
            {assignments.length === 0 ? <p>La rutina puede crearse sin sesiones.</p> : (
              <ul className="space-y-2">
                {assignments.map((item, index) => (
                  <li className="flex flex-wrap items-center justify-between gap-2 rounded-box border border-base-300 p-3" key={`${item.session.id}-${item.day}`}>
                    <span><strong>{weekdays[item.day]}:</strong> {item.session.name}</span>
                    <button className="btn btn-error btn-sm" type="button" onClick={() => setAssignments((current) => current.filter((_, itemIndex) => itemIndex !== index))}>Quitar {item.session.name} del {weekdays[item.day].toLowerCase()}</button>
                  </li>
                ))}
              </ul>
            )}
          </fieldset>
          <button className="btn btn-primary" type="submit">Crear rutina</button>
        </div>
      </form>

      <div aria-busy={loading}>
        <h3 className="mb-3 text-xl font-semibold">Mis rutinas</h3>
        {loading ? <p role="status">Cargando rutinas…</p> : routines.length === 0 ? <p className="alert">Todavía no creaste rutinas.</p> : (
          <ul className="grid gap-3 md:grid-cols-2">
            {routines.map((routine) => (
              <li className="card border border-base-300" key={routine.id}>
                <div className="card-body">
                  <h4 className="card-title">{routine.name}</h4>
                  {routine.description && <p>{routine.description}</p>}
                  <div className="card-actions">
                    <button className="btn btn-sm" type="button" onClick={() => void inspect(routine.id)}>Ver {routine.name}</button>
                    <button className="btn btn-error btn-sm" type="button" onClick={() => void remove(routine)}>Eliminar {routine.name}</button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {selected && (
        <article className="card border border-primary" aria-labelledby={`routine-detail-${selected.id}`}>
          <div className="card-body gap-4">
            <h3 className="card-title" id={`routine-detail-${selected.id}`}>Detalle: {selected.name}</h3>
            {selected.description && <p>{selected.description}</p>}
            {selected.sessions.length === 0 ? <p>Esta rutina no tiene sesiones.</p> : selected.sessions.map((item, index) => (
              <section aria-labelledby={`routine-${selected.id}-session-${index}`} className="space-y-2" key={`${item.day}-${item.session.id}`}>
                <h4 className="text-xl font-semibold" id={`routine-${selected.id}-session-${index}`}>{weekdays[item.day]}</h4>
                <SessionDetailCard session={item.session} />
              </section>
            ))}
          </div>
        </article>
      )}
    </section>
  )
}
