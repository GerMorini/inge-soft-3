import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { SessionsView } from './SessionsView'

const exercises = [{ id: 1, name: 'Sentadilla' }, { id: 2, name: 'Plancha' }]

function jsonResponse(body: unknown, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }))
}

describe('SessionsView', () => {
  beforeEach(() => {
    sessionStorage.setItem('accessToken', 'token')
    vi.restoreAllMocks()
  })

  it('blocks duplicate exercises, reorders by buttons and submits exact order', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (init?.method === 'POST') return jsonResponse({ id: 4, name: 'Piernas', exercises: [] }, 201)
      if (path === '/api/exercises') return jsonResponse(exercises)
      return jsonResponse([])
    })
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    render(<SessionsView onUnauthenticated={vi.fn()} />)
    await screen.findByText('Todavía no creaste sesiones.')
    const select = screen.getByLabelText('Ejercicio disponible')
    await user.selectOptions(select, '1')
    await user.click(screen.getByRole('button', { name: 'Agregar ejercicio' }))
    await user.selectOptions(select, '1')
    await user.click(screen.getByRole('button', { name: 'Agregar ejercicio' }))
    expect(screen.getByText('Ese ejercicio ya está incluido en la sesión.')).toBeInTheDocument()
    await user.selectOptions(select, '2')
    await user.click(screen.getByRole('button', { name: 'Agregar ejercicio' }))
    await user.click(screen.getByRole('button', { name: 'Subir Plancha' }))
    expect(screen.getByText('1. Plancha')).toBeInTheDocument()
    await user.type(screen.getByLabelText('Nombre *'), 'Piernas')
    await user.clear(screen.getByLabelText('Series de Plancha'))
    await user.type(screen.getByLabelText('Series de Plancha'), '3')
    await user.clear(screen.getByLabelText('Repeticiones de Plancha'))
    await user.type(screen.getByLabelText('Repeticiones de Plancha'), '10')
    await user.click(screen.getByRole('button', { name: 'Crear sesión' }))
    await screen.findByText('Sesión Piernas creada.')
    const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST')
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({
      name: 'Piernas', description: '', exercises: [
        { exerciseId: 2, series: 3, repetitions: 10, order: 1 },
        { exerciseId: 1, series: 0, repetitions: 0, order: 2 },
      ],
    })
  })

  it('renders full detail and respects deletion confirmation', async () => {
    const summary = { id: 5, name: 'Fuerza' }
    const detail = { id: 5, name: 'Fuerza', exercises: [{ exercise: { id: 1, name: 'Sentadilla', videoUrl: 'https://example.com/video' }, series: 4, repetitions: 8, order: 1 }] }
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (init?.method === 'DELETE') return Promise.resolve(new Response(null, { status: 204 }))
      if (path === '/api/exercises') return jsonResponse(exercises)
      if (path === '/api/sessions/5') return jsonResponse(detail)
      return jsonResponse([summary])
    })
    vi.stubGlobal('fetch', fetchMock)
    const confirm = vi.spyOn(window, 'confirm').mockReturnValueOnce(false).mockReturnValueOnce(true)
    const user = userEvent.setup()
    render(<SessionsView onUnauthenticated={vi.fn()} />)
    await user.click(await screen.findByRole('button', { name: 'Ver Fuerza' }))
    expect(await screen.findByText('4 series · 8 repeticiones')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Abrir video de Sentadilla' })).toHaveAttribute('rel', 'noreferrer')
    const remove = screen.getByRole('button', { name: 'Eliminar Fuerza' })
    await user.click(remove)
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'DELETE')).toBe(false)
    await user.click(remove)
    expect(await screen.findByText('Sesión Fuerza eliminada.')).toBeInTheDocument()
    expect(confirm).toHaveBeenCalledTimes(2)
  })

  it('replaces a complete edited composition with derived order', async () => {
    const summary = { id: 6, name: 'Mixta' }
    const detail = {
      id: 6, name: 'Mixta', description: 'Original', exercises: [
        { exercise: exercises[0], series: 3, repetitions: 8, order: 1 },
        { exercise: exercises[1], series: 2, repetitions: 30, order: 2 },
      ],
    }
    const updated = { ...detail, name: 'Mixta nueva', description: undefined, exercises: [detail.exercises[1]] }
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (init?.method === 'PUT') return jsonResponse(updated)
      if (path === '/api/exercises') return jsonResponse(exercises)
      if (path === '/api/sessions/6') return jsonResponse(detail)
      return jsonResponse([summary])
    })
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    render(<SessionsView onUnauthenticated={vi.fn()} />)

    await user.click(await screen.findByRole('button', { name: 'Editar Mixta' }))
    const name = screen.getByLabelText('Nombre *')
    await user.clear(name)
    await user.type(name, 'Mixta nueva')
    await user.clear(screen.getByLabelText('Descripción'))
    await user.click(screen.getByRole('button', { name: 'Quitar Sentadilla' }))
    await user.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('Sesión Mixta nueva actualizada.')).toBeInTheDocument()
    const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'PUT')
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({
      name: 'Mixta nueva', description: '', exercises: [
        { exerciseId: 2, series: 2, repetitions: 30, order: 1 },
      ],
    })
  })
})
