import { useEffect, useState, type FormEvent } from 'react'
import { ApiError } from '../auth/types'
import { createSession, deleteSession, getSession, listExercises, listSessions } from './api'
import type { Exercise, FieldErrors, SessionDetail, SessionSummary } from './types'

interface SessionsViewProps {
  onUnauthenticated: () => void
}

interface SelectedExerciseRow {
  exercise: Exercise
  series: number
  repetitions: number
}

export function SessionsView({ onUnauthenticated }: SessionsViewProps) {
  const [sessions, setSessions] = useState<SessionSummary[]>([])
  const [exercises, setExercises] = useState<Exercise[]>([])
  const [selected, setSelected] = useState<SessionDetail | null>(null)
  const [composition, setComposition] = useState<SelectedExerciseRow[]>([])
  const [exerciseID, setExerciseID] = useState('')
  const [form, setForm] = useState({ name: '', description: '' })
  const [fields, setFields] = useState<FieldErrors>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [sessionItems, exerciseItems] = await Promise.all([
        listSessions(onUnauthenticated),
        listExercises(onUnauthenticated),
      ])
      setSessions(sessionItems)
      setExercises(exerciseItems)
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudieron cargar las sesiones.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  function addExercise() {
    const exercise = exercises.find((item) => item.id === Number(exerciseID))
    if (!exercise) return
    if (composition.some((item) => item.exercise.id === exercise.id)) {
      setError('Ese ejercicio ya está incluido en la sesión.')
      return
    }
    setComposition((current) => [...current, { exercise, series: 0, repetitions: 0 }])
    setExerciseID('')
    setError('')
  }

  function move(index: number, direction: -1 | 1) {
    const destination = index + direction
    if (destination < 0 || destination >= composition.length) return
    setComposition((current) => {
      const next = [...current]
      ;[next[index], next[destination]] = [next[destination], next[index]]
      return next
    })
  }

  function updateQuantity(index: number, field: 'series' | 'repetitions', value: number) {
    setComposition((current) => current.map((item, itemIndex) => itemIndex === index ? { ...item, [field]: value } : item))
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    const nextFields: FieldErrors = {}
    if (!form.name) nextFields.name = ['Este campo es obligatorio.']
    else if (form.name.trim() !== form.name || /\s{2,}|[\t\n\r]/.test(form.name) || form.name.length > 100) nextFields.name = ['Usá hasta 100 caracteres y un espacio entre palabras.']
    if (form.description && (form.description.trim() !== form.description || /\s{2,}|[\t\n\r]/.test(form.description) || form.description.length > 500)) nextFields.description = ['Usá hasta 500 caracteres y un espacio entre palabras.']
    composition.forEach((item, index) => {
      if (!Number.isInteger(item.series) || item.series < 0) nextFields[`exercises.${index}.series`] = ['Debe ser un entero no negativo.']
      if (!Number.isInteger(item.repetitions) || item.repetitions < 0) nextFields[`exercises.${index}.repetitions`] = ['Debe ser un entero no negativo.']
    })
    if (Object.keys(nextFields).length > 0) {
      setFields(nextFields)
      setError('Revisá los campos indicados.')
      return
    }
    setFields({})
    setError('')
    try {
      const created = await createSession({
        ...form,
        exercises: composition.map((item, index) => ({
          exerciseId: item.exercise.id,
          series: item.series,
          repetitions: item.repetitions,
          order: index + 1,
        })),
      }, onUnauthenticated)
      setSessions((current) => [...current, created].sort((a, b) => a.id - b.id))
      setSelected(created)
      setComposition([])
      setForm({ name: '', description: '' })
      setMessage(`Sesión ${created.name} creada.`)
    } catch (reason) {
      if (reason instanceof ApiError) setFields(reason.body.fields ?? {})
      if (!(reason instanceof ApiError && reason.status === 401)) setError(reason instanceof Error ? reason.message : 'No se pudo crear la sesión.')
    }
  }

  async function inspect(id: number) {
    setError('')
    try {
      setSelected(await getSession(id, onUnauthenticated))
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudo cargar la sesión.')
    }
  }

  async function remove(session: SessionSummary) {
    if (!window.confirm(`¿Eliminar la sesión ${session.name}?`)) return
    setError('')
    try {
      await deleteSession(session.id, onUnauthenticated)
      setSessions((current) => current.filter((item) => item.id !== session.id))
      setSelected((current) => current?.id === session.id ? null : current)
      setMessage(`Sesión ${session.name} eliminada.`)
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudo eliminar la sesión.')
    }
  }

  return (
    <section aria-labelledby="sessions-title" className="space-y-6">
      <h2 className="text-2xl font-bold" id="sessions-title">Sesiones</h2>
      {message && <p className="alert alert-success" role="status">{message}</p>}
      {error && <p className="alert alert-error" role="alert" tabIndex={-1}>{error}</p>}

      <form aria-label="Crear sesión" className="card bg-base-200" onSubmit={submit} noValidate>
        <div className="card-body gap-4">
          <h3 className="card-title">Crear sesión</h3>
          <label className="fieldset">
            <span className="fieldset-legend">Nombre *</span>
            <input className={`input w-full ${fields.name ? 'input-error' : ''}`} value={form.name} required aria-invalid={fields.name ? 'true' : undefined} aria-describedby={fields.name ? 'session-name-error' : undefined} onChange={(event) => setForm({ ...form, name: event.target.value })} />
            {fields.name && <span className="text-error" id="session-name-error">{fields.name.join(' ')}</span>}
          </label>
          <label className="fieldset">
            <span className="fieldset-legend">Descripción</span>
            <textarea className={`textarea w-full ${fields.description ? 'textarea-error' : ''}`} value={form.description} aria-invalid={fields.description ? 'true' : undefined} aria-describedby={fields.description ? 'session-description-error' : undefined} onChange={(event) => setForm({ ...form, description: event.target.value })} />
            {fields.description && <span className="text-error" id="session-description-error">{fields.description.join(' ')}</span>}
          </label>

          <fieldset className="fieldset rounded-box border border-base-300 p-4">
            <legend className="fieldset-legend px-2">Ejercicios de la sesión</legend>
            <div className="flex flex-col gap-2 sm:flex-row">
              <label className="sr-only" htmlFor="session-exercise">Ejercicio disponible</label>
              <select className="select flex-1" id="session-exercise" value={exerciseID} onChange={(event) => setExerciseID(event.target.value)}>
                <option value="">Elegí un ejercicio</option>
                {exercises.map((exercise) => <option key={exercise.id} value={exercise.id}>{exercise.name}</option>)}
              </select>
              <button className="btn" type="button" disabled={!exerciseID} onClick={addExercise}>Agregar ejercicio</button>
            </div>
            {composition.length === 0 ? <p>La sesión puede crearse sin ejercicios.</p> : (
              <ol className="space-y-3">
                {composition.map((item, index) => (
                  <li className="rounded-box border border-base-300 p-3" key={item.exercise.id}>
                    <p className="font-semibold">{index + 1}. {item.exercise.name}</p>
                    <div className="grid gap-3 sm:grid-cols-2">
                      <label className="fieldset">
                        <span className="fieldset-legend">Series de {item.exercise.name}</span>
                        <input className="input" type="number" min="0" step="1" value={item.series} aria-invalid={fields[`exercises.${index}.series`] ? 'true' : undefined} onChange={(event) => updateQuantity(index, 'series', event.target.valueAsNumber)} />
                        {fields[`exercises.${index}.series`] && <span className="text-error">{fields[`exercises.${index}.series`].join(' ')}</span>}
                      </label>
                      <label className="fieldset">
                        <span className="fieldset-legend">Repeticiones de {item.exercise.name}</span>
                        <input className="input" type="number" min="0" step="1" value={item.repetitions} aria-invalid={fields[`exercises.${index}.repetitions`] ? 'true' : undefined} onChange={(event) => updateQuantity(index, 'repetitions', event.target.valueAsNumber)} />
                        {fields[`exercises.${index}.repetitions`] && <span className="text-error">{fields[`exercises.${index}.repetitions`].join(' ')}</span>}
                      </label>
                    </div>
                    <div className="mt-2 flex flex-wrap gap-2">
                      <button className="btn btn-sm" type="button" disabled={index === 0} onClick={() => move(index, -1)}>Subir {item.exercise.name}</button>
                      <button className="btn btn-sm" type="button" disabled={index === composition.length - 1} onClick={() => move(index, 1)}>Bajar {item.exercise.name}</button>
                      <button className="btn btn-error btn-sm" type="button" onClick={() => setComposition((current) => current.filter((_, itemIndex) => itemIndex !== index))}>Quitar {item.exercise.name}</button>
                    </div>
                  </li>
                ))}
              </ol>
            )}
          </fieldset>
          <button className="btn btn-primary" type="submit">Crear sesión</button>
        </div>
      </form>

      <div aria-busy={loading}>
        <h3 className="mb-3 text-xl font-semibold">Mis sesiones</h3>
        {loading ? <p role="status">Cargando sesiones…</p> : sessions.length === 0 ? <p className="alert">Todavía no creaste sesiones.</p> : (
          <ul className="grid gap-3 md:grid-cols-2">
            {sessions.map((session) => (
              <li className="card border border-base-300" key={session.id}>
                <div className="card-body">
                  <h4 className="card-title">{session.name}</h4>
                  {session.description && <p>{session.description}</p>}
                  <div className="card-actions">
                    <button className="btn btn-sm" type="button" onClick={() => void inspect(session.id)}>Ver {session.name}</button>
                    <button className="btn btn-error btn-sm" type="button" onClick={() => void remove(session)}>Eliminar {session.name}</button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {selected && <SessionDetailCard session={selected} />}
    </section>
  )
}

export function SessionDetailCard({ session }: { session: SessionDetail }) {
  return (
    <article className="card border border-primary" aria-labelledby={`session-detail-${session.id}`}>
      <div className="card-body">
        <h3 className="card-title" id={`session-detail-${session.id}`}>Detalle: {session.name}</h3>
        {session.description && <p>{session.description}</p>}
        {session.exercises.length === 0 ? <p>Esta sesión no tiene ejercicios.</p> : (
          <ol className="space-y-3">
            {session.exercises.map((item) => (
              <li className="rounded-box bg-base-200 p-3" key={item.exercise.id}>
                <h4 className="font-semibold">{item.order}. {item.exercise.name}</h4>
                {item.exercise.description && <p>{item.exercise.description}</p>}
                <p>{item.series} series · {item.repetitions} repeticiones</p>
                <div className="flex flex-wrap gap-3">
                  {item.exercise.imageUrl && <a className="link" href={item.exercise.imageUrl} target="_blank" rel="noreferrer">Abrir imagen de {item.exercise.name}</a>}
                  {item.exercise.videoUrl && <a className="link" href={item.exercise.videoUrl} target="_blank" rel="noreferrer">Abrir video de {item.exercise.name}</a>}
                </div>
              </li>
            ))}
          </ol>
        )}
      </div>
    </article>
  )
}
