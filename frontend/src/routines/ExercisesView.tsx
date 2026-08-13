import { useEffect, useState, type FormEvent } from 'react'
import { ApiError } from '../auth/types'
import { createExercise, deleteExercise, getExercise, listExercises } from './api'
import type { Exercise, FieldErrors } from './types'

interface ExercisesViewProps {
  onUnauthenticated: () => void
}

function validateText(value: string, maximum: number, required: boolean): string[] {
  const errors: string[] = []
  if (required && value === '') errors.push('Este campo es obligatorio.')
  if (value.length > maximum) errors.push(`No puede superar ${maximum} caracteres.`)
  if (value && (value.trim() !== value || /\s{2,}|[\t\n\r]/.test(value))) {
    errors.push('Usá solo un espacio entre palabras, sin espacios al inicio o final.')
  }
  return errors
}

function validateURL(value: string): string[] {
  if (!value) return []
  if (value.length > 2048) return ['No puede superar 2048 caracteres.']
  try {
    const parsed = new URL(value)
    if (!['http:', 'https:'].includes(parsed.protocol) || !parsed.hostname || /\s/.test(value)) throw new Error()
  } catch {
    return ['Ingresá una URL HTTP o HTTPS válida.']
  }
  return []
}

export function ExercisesView({ onUnauthenticated }: ExercisesViewProps) {
  const [exercises, setExercises] = useState<Exercise[]>([])
  const [selected, setSelected] = useState<Exercise | null>(null)
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [fields, setFields] = useState<FieldErrors>({})
  const [form, setForm] = useState({ name: '', description: '', imageUrl: '', videoUrl: '' })

  async function load() {
    setLoading(true)
    setError('')
    try {
      setExercises(await listExercises(onUnauthenticated))
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudieron cargar los ejercicios.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function submit(event: FormEvent) {
    event.preventDefault()
    const nextFields: FieldErrors = {
      name: validateText(form.name, 100, true),
      description: validateText(form.description, 500, false),
      imageUrl: validateURL(form.imageUrl),
      videoUrl: validateURL(form.videoUrl),
    }
    for (const key of Object.keys(nextFields)) if (nextFields[key].length === 0) delete nextFields[key]
    if (Object.keys(nextFields).length > 0) {
      setFields(nextFields)
      setError('Revisá los campos indicados.')
      return
    }
    setError('')
    setFields({})
    try {
      const created = await createExercise(form, onUnauthenticated)
      setExercises((current) => [...current, created].sort((a, b) => a.id - b.id))
      setSelected(created)
      setForm({ name: '', description: '', imageUrl: '', videoUrl: '' })
      setMessage(`Ejercicio ${created.name} creado.`)
    } catch (reason) {
      if (reason instanceof ApiError) setFields(reason.body.fields ?? {})
      if (!(reason instanceof ApiError && reason.status === 401)) setError(reason instanceof Error ? reason.message : 'No se pudo crear el ejercicio.')
    }
  }

  async function inspect(id: number) {
    setError('')
    try {
      setSelected(await getExercise(id, onUnauthenticated))
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudo cargar el ejercicio.')
    }
  }

  async function remove(exercise: Exercise) {
    if (!window.confirm(`¿Eliminar el ejercicio ${exercise.name}?`)) return
    setError('')
    try {
      await deleteExercise(exercise.id, onUnauthenticated)
      setSelected((current) => current?.id === exercise.id ? null : current)
      setExercises((current) => current.filter((item) => item.id !== exercise.id))
      setMessage(`Ejercicio ${exercise.name} eliminado.`)
    } catch (reason) {
      if (!(reason instanceof ApiError && reason.status === 401)) setError('No se pudo eliminar el ejercicio.')
    }
  }

  return (
    <section aria-labelledby="exercises-title" className="space-y-6">
      <h2 className="text-2xl font-bold" id="exercises-title">Ejercicios</h2>
      {message && <p className="alert alert-success" role="status">{message}</p>}
      {error && <p className="alert alert-error" role="alert" tabIndex={-1}>{error}</p>}

      <form aria-label="Crear ejercicio" className="card bg-base-200" onSubmit={submit} noValidate>
        <div className="card-body grid gap-4 md:grid-cols-2">
          <h3 className="card-title md:col-span-2">Crear ejercicio</h3>
          <Field label="Nombre" name="exercise-name" value={form.name} errors={fields.name} required onChange={(name) => setForm({ ...form, name })} />
          <Field label="Descripción" name="exercise-description" value={form.description} errors={fields.description} onChange={(description) => setForm({ ...form, description })} />
          <Field label="URL de imagen" name="exercise-image-url" value={form.imageUrl} errors={fields.imageUrl} inputMode="url" onChange={(imageUrl) => setForm({ ...form, imageUrl })} />
          <Field label="URL de video" name="exercise-video-url" value={form.videoUrl} errors={fields.videoUrl} inputMode="url" onChange={(videoUrl) => setForm({ ...form, videoUrl })} />
          <button className="btn btn-primary md:col-span-2" type="submit">Crear ejercicio</button>
        </div>
      </form>

      <div aria-busy={loading}>
        <h3 className="mb-3 text-xl font-semibold">Mis ejercicios</h3>
        {loading ? <p role="status">Cargando ejercicios…</p> : exercises.length === 0 ? (
          <p className="alert">Todavía no creaste ejercicios.</p>
        ) : (
          <ul className="grid gap-3 md:grid-cols-2">
            {exercises.map((exercise) => (
              <li className="card border border-base-300" key={exercise.id}>
                <div className="card-body">
                  <h4 className="card-title">{exercise.name}</h4>
                  {exercise.description && <p>{exercise.description}</p>}
                  <div className="card-actions">
                    <button className="btn btn-sm" type="button" onClick={() => void inspect(exercise.id)}>Ver {exercise.name}</button>
                    <button className="btn btn-error btn-sm" type="button" onClick={() => void remove(exercise)}>Eliminar {exercise.name}</button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {selected && (
        <article className="card border border-primary" aria-labelledby="exercise-detail-title">
          <div className="card-body">
            <h3 className="card-title" id="exercise-detail-title">Detalle: {selected.name}</h3>
            {selected.description && <p>{selected.description}</p>}
            <div className="flex flex-wrap gap-3">
              {selected.imageUrl && <a className="link link-primary" href={selected.imageUrl} target="_blank" rel="noreferrer">Abrir imagen de {selected.name}</a>}
              {selected.videoUrl && <a className="link link-primary" href={selected.videoUrl} target="_blank" rel="noreferrer">Abrir video de {selected.name}</a>}
            </div>
          </div>
        </article>
      )}
    </section>
  )
}

interface FieldProps {
  label: string
  name: string
  value: string
  errors?: string[]
  required?: boolean
  inputMode?: 'url'
  onChange: (value: string) => void
}

function Field({ label, name, value, errors, required, inputMode, onChange }: FieldProps) {
  const errorID = `${name}-error`
  return (
    <div className="fieldset">
      <label className="fieldset-legend" htmlFor={name}>{label}{required ? ' *' : ''}</label>
      <input className={`input w-full ${errors ? 'input-error' : ''}`} id={name} value={value} required={required} inputMode={inputMode} aria-invalid={errors ? 'true' : undefined} aria-describedby={errors ? errorID : undefined} onChange={(event) => onChange(event.target.value)} />
      {errors && <p className="text-error" id={errorID}>{errors.join(' ')}</p>}
    </div>
  )
}
