import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import App from './App'

function jsonResponse(body: unknown) {
  return Promise.resolve(new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  }))
}

describe('App workspace navigation', () => {
  beforeEach(() => {
    sessionStorage.setItem('accessToken', browserToken(Date.now() + 60_000))
    vi.restoreAllMocks()
  })

  it('keeps a dirty edit when discard is rejected and switches without PUT when accepted', async () => {
    const routine = { id: 4, name: 'Semana', sessions: [] }
    const fetchMock = vi.fn((input: RequestInfo | URL, _init?: RequestInit) => {
      const path = String(input)
      if (path === '/api/auth/me') return jsonResponse({ id: 1, username: 'ada' })
      if (path === '/api/routines/4') return jsonResponse(routine)
      if (path === '/api/routines') return jsonResponse([{ id: 4, name: 'Semana' }])
      return jsonResponse([])
    })
    vi.stubGlobal('fetch', fetchMock)
    const confirm = vi.spyOn(window, 'confirm').mockReturnValueOnce(false).mockReturnValueOnce(true)
    const user = userEvent.setup()
    render(<App />)

    await user.click(await screen.findByRole('button', { name: 'Editar Semana' }))
    const name = screen.getByLabelText('Nombre *')
    await user.clear(name)
    await user.type(name, 'Semana nueva')
    await user.click(screen.getByRole('tab', { name: 'Sesiones' }))
    expect(screen.getByRole('tab', { name: 'Rutinas' })).toHaveAttribute('aria-selected', 'true')
    expect(name).toHaveValue('Semana nueva')

    await user.click(screen.getByRole('tab', { name: 'Sesiones' }))
    expect(await screen.findByText('Todavía no creaste sesiones.')).toBeInTheDocument()
    expect(confirm).toHaveBeenCalledTimes(2)
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'PUT')).toBe(false)
  })
})

function browserToken(expiresAt: number): string {
  const payload = btoa(JSON.stringify({ exp: Math.floor(expiresAt / 1000) }))
  return `header.${payload}.signature`
}
