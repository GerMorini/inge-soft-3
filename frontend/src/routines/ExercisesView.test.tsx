import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { ExercisesView } from './ExercisesView'

function jsonResponse(body: unknown, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }))
}

describe('ExercisesView', () => {
  beforeEach(() => {
    sessionStorage.setItem('accessToken', 'token')
    vi.restoreAllMocks()
  })

  it('shows loading, empty state and field validation without clearing input', async () => {
    vi.stubGlobal('fetch', vi.fn(() => jsonResponse([])))
    const user = userEvent.setup()
    render(<ExercisesView onUnauthenticated={vi.fn()} />)
    expect(screen.getByText('Cargando ejercicios…')).toBeInTheDocument()
    expect(await screen.findByText('Todavía no creaste ejercicios.')).toBeInTheDocument()
    await user.type(screen.getByLabelText('Nombre *'), ' Peso  muerto ')
    await user.type(screen.getByLabelText('URL de imagen'), 'imagen-relativa')
    await user.click(screen.getByRole('button', { name: 'Crear ejercicio' }))
    expect(screen.getByText('Revisá los campos indicados.')).toBeInTheDocument()
    expect(screen.getByLabelText('Nombre *')).toHaveValue(' Peso  muerto ')
    expect(fetch).toHaveBeenCalledTimes(1)
  })

  it('creates an exercise and renders safe optional links', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return jsonResponse({ id: 2, name: 'Plancha', imageUrl: 'https://example.com/image', videoUrl: 'https://example.com/video' }, 201)
      return jsonResponse([])
    })
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    render(<ExercisesView onUnauthenticated={vi.fn()} />)
    await screen.findByText('Todavía no creaste ejercicios.')
    await user.type(screen.getByLabelText('Nombre *'), 'Plancha')
    await user.type(screen.getByLabelText('URL de imagen'), 'https://example.com/image')
    await user.type(screen.getByLabelText('URL de video'), 'https://example.com/video')
    await user.click(screen.getByRole('button', { name: 'Crear ejercicio' }))
    expect(await screen.findByText('Ejercicio Plancha creado.')).toBeInTheDocument()
    const imageLink = screen.getByRole('link', { name: 'Abrir imagen de Plancha' })
    expect(imageLink).toHaveAttribute('target', '_blank')
    expect(imageLink).toHaveAttribute('rel', 'noreferrer')
    const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST')
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({ name: 'Plancha', description: '', imageUrl: 'https://example.com/image', videoUrl: 'https://example.com/video' })
  })

  it('cancels deletion, confirms deletion and handles lost authentication', async () => {
    const exercise = { id: 3, name: 'Remo' }
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'DELETE') return Promise.resolve(new Response(null, { status: 204 }))
      return jsonResponse([exercise])
    })
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    const confirm = vi.spyOn(window, 'confirm').mockReturnValueOnce(false).mockReturnValueOnce(true)
    render(<ExercisesView onUnauthenticated={vi.fn()} />)
    const remove = await screen.findByRole('button', { name: 'Eliminar Remo' })
    await user.click(remove)
    expect(confirm).toHaveBeenCalled()
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'DELETE')).toBe(false)
    await user.click(remove)
    expect(await screen.findByText('Ejercicio Remo eliminado.')).toBeInTheDocument()

    const onUnauthenticated = vi.fn()
    vi.stubGlobal('fetch', vi.fn(() => jsonResponse({ error: { code: 'invalid_token', message: 'Vencido' } }, 401)))
    render(<ExercisesView onUnauthenticated={onUnauthenticated} />)
    await waitFor(() => expect(onUnauthenticated).toHaveBeenCalled())
    expect(sessionStorage.getItem('accessToken')).toBeNull()
  })
})
