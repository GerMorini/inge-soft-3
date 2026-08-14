import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { RoutinesView } from './RoutinesView'

function jsonResponse(body: unknown, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }))
}

describe('RoutinesView', () => {
  beforeEach(() => {
    sessionStorage.setItem('accessToken', 'token')
    vi.restoreAllMocks()
  })

  it('allows a session on different days, blocks duplicate pairs and sends ISO days', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (init?.method === 'POST') return jsonResponse({ id: 7, name: 'Semana A', sessions: [] }, 201)
      if (path === '/api/sessions') return jsonResponse([{ id: 3, name: 'Piernas' }])
      return jsonResponse([])
    })
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    render(<RoutinesView onUnauthenticated={vi.fn()} />)
    await screen.findByText('Todavía no creaste rutinas.')
    const session = screen.getByLabelText('Sesión')
    const day = screen.getByLabelText('Día')
    await user.selectOptions(session, '3')
    await user.selectOptions(day, '1')
    await user.click(screen.getByRole('button', { name: 'Agregar sesión' }))
    await user.selectOptions(session, '3')
    await user.selectOptions(day, '7')
    await user.click(screen.getByRole('button', { name: 'Agregar sesión' }))
    await user.selectOptions(session, '3')
    await user.selectOptions(day, '7')
    await user.click(screen.getByRole('button', { name: 'Agregar sesión' }))
    expect(screen.getByText('Piernas ya está asignada al domingo.')).toBeInTheDocument()
    await user.type(screen.getByLabelText('Nombre *'), 'Semana A')
    await user.click(screen.getByRole('button', { name: 'Crear rutina' }))
    await screen.findByText('Rutina Semana A creada.')
    const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST')
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({
      name: 'Semana A', description: '', sessions: [{ sessionId: 3, day: 1 }, { sessionId: 3, day: 7 }],
    })
  })

  it('renders complete nested routine details and deletion outcomes', async () => {
    const summary = { id: 8, name: 'Semana completa' }
    const detail = {
      id: 8, name: 'Semana completa', sessions: [{ day: 2, session: {
        id: 3, name: 'Tren superior', exercises: [{ exercise: { id: 4, name: 'Remo', imageUrl: 'https://example.com/remo' }, series: 3, repetitions: 12, order: 1 }],
      } }],
    }
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (init?.method === 'DELETE') return Promise.resolve(new Response(JSON.stringify({ error: { code: 'internal_error', message: 'Error' } }), { status: 500 }))
      if (path === '/api/sessions') return jsonResponse([{ id: 3, name: 'Tren superior' }])
      if (path === '/api/routines/8') return jsonResponse(detail)
      return jsonResponse([summary])
    })
    vi.stubGlobal('fetch', fetchMock)
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const user = userEvent.setup()
    render(<RoutinesView onUnauthenticated={vi.fn()} />)
    await user.click(await screen.findByRole('button', { name: 'Ver Semana completa' }))
    expect(await screen.findByRole('heading', { name: 'Martes' })).toBeInTheDocument()
    expect(screen.getByText('3 series · 12 repeticiones')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Abrir imagen de Remo' })).toHaveAttribute('target', '_blank')
    await user.click(screen.getByRole('button', { name: 'Eliminar Semana completa' }))
    expect(await screen.findByText('No se pudo eliminar la rutina.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Eliminar Semana completa' })).toBeInTheDocument()
  })

  it('prefills and replaces all edited routine assignments', async () => {
    const session = { id: 3, name: 'Piernas', exercises: [] }
    const summary = { id: 10, name: 'Semana A' }
    const detail = { id: 10, name: 'Semana A', description: 'Vieja', sessions: [{ day: 1, session }] }
    const updated = { id: 10, name: 'Semana B', sessions: [{ day: 7, session }] }
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (init?.method === 'PUT') return jsonResponse(updated)
      if (path === '/api/sessions') return jsonResponse([session])
      if (path === '/api/routines/10') return jsonResponse(detail)
      return jsonResponse([summary])
    })
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    render(<RoutinesView onUnauthenticated={vi.fn()} />)

    await user.click(await screen.findByRole('button', { name: 'Editar Semana A' }))
    const name = screen.getByLabelText('Nombre *')
    await user.clear(name)
    await user.type(name, 'Semana B')
    await user.clear(screen.getByLabelText('Descripción'))
    await user.click(screen.getByRole('button', { name: 'Quitar Piernas del lunes' }))
    await user.selectOptions(screen.getByLabelText('Sesión'), '3')
    await user.selectOptions(screen.getByLabelText('Día'), '7')
    await user.click(screen.getByRole('button', { name: 'Agregar sesión' }))
    await user.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('Rutina Semana B actualizada.')).toBeInTheDocument()
    const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'PUT')
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({
      name: 'Semana B', description: '', sessions: [{ sessionId: 3, day: 7 }],
    })
  })
})
