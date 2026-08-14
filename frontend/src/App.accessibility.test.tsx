import axe from 'axe-core'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import App from './App'

describe('App accessibility', () => {
  beforeEach(() => {
    sessionStorage.clear()
    vi.restoreAllMocks()
  })
  it('has no automated accessibility violations in the registration view', async () => {
    const { container } = render(<App />)
    const result = await axe.run(container, {
      rules: {
        'color-contrast': { enabled: false },
      },
    })
    expect(result.violations).toEqual([])
  })

  it('has no automated violations with visible passwords and mismatch feedback', async () => {
    const user = userEvent.setup()
    const { container } = render(<App />)

    await user.type(screen.getByLabelText('Contraseña'), 'Segura!@123')
    await user.type(screen.getByLabelText('Confirmar contraseña'), 'Distinta!123')
    await user.click(screen.getByRole('button', { name: 'Ver contraseña ingresada' }))
    await user.click(screen.getByRole('button', { name: 'Ver contraseña de confirmación' }))
    fireEvent.submit(screen.getByRole('form', { name: 'Registro' }))

    expect(await screen.findByText('Las contraseñas deben coincidir.')).toBeInTheDocument()
    const result = await axe.run(container, {
      rules: {
        'color-contrast': { enabled: false },
      },
    })
    expect(result.violations).toEqual([])
  })

  it('keeps authenticated tabs, forms and deletion controls accessible', async () => {
    const user = userEvent.setup()
    sessionStorage.setItem('accessToken', browserToken(Date.now() + 60_000))
    vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
      const path = String(input)
      const body = path === '/api/auth/me'
        ? { id: 42, username: 'ada_01' }
        : path === '/api/routines'
          ? [{ id: 7, name: 'Semana A' }]
          : path === '/api/sessions'
            ? [{ id: 5, name: 'Piernas' }]
            : path === '/api/exercises'
              ? [{ id: 3, name: 'Sentadilla' }]
              : []
      return Promise.resolve(new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    })
    const { container } = render(<App />)
    expect(await screen.findByRole('button', { name: 'Eliminar Semana A' })).toBeInTheDocument()

    const routinesTab = screen.getByRole('tab', { name: 'Rutinas' })
    routinesTab.focus()
    await user.keyboard('{ArrowRight}')
    expect(screen.getByRole('tab', { name: 'Sesiones' })).toHaveFocus()
    expect(await screen.findByRole('button', { name: 'Eliminar Piernas' })).toBeInTheDocument()
    await user.click(screen.getByRole('tab', { name: 'Ejercicios' }))
    expect(await screen.findByRole('button', { name: 'Eliminar Sentadilla' })).toBeInTheDocument()

    await waitFor(async () => {
      const result = await axe.run(container, { rules: { 'color-contrast': { enabled: false } } })
      expect(result.violations).toEqual([])
    })
  })

  it('keeps every prefilled edit mode accessible and moves focus to its heading', async () => {
    const user = userEvent.setup()
    sessionStorage.setItem('accessToken', browserToken(Date.now() + 60_000))
    vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
      const path = String(input)
      const body = path === '/api/auth/me'
        ? { id: 42, username: 'ada_01' }
        : path === '/api/routines'
          ? [{ id: 7, name: 'Semana A' }]
          : path === '/api/routines/7'
            ? { id: 7, name: 'Semana A', sessions: [] }
            : path === '/api/sessions'
              ? [{ id: 5, name: 'Piernas' }]
              : path === '/api/sessions/5'
                ? { id: 5, name: 'Piernas', exercises: [] }
                : path === '/api/exercises'
                  ? [{ id: 3, name: 'Sentadilla' }]
                  : path === '/api/exercises/3'
                    ? { id: 3, name: 'Sentadilla' }
                    : []
      return Promise.resolve(new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    })
    const { container } = render(<App />)

    await user.click(await screen.findByRole('button', { name: 'Editar Semana A' }))
    expect(screen.getByRole('heading', { name: 'Editar rutina' })).toHaveFocus()
    expect((await axe.run(container, { rules: { 'color-contrast': { enabled: false } } })).violations).toEqual([])

    await user.click(screen.getByRole('tab', { name: 'Sesiones' }))
    await user.click(await screen.findByRole('button', { name: 'Editar Piernas' }))
    expect(screen.getByRole('heading', { name: 'Editar sesión' })).toHaveFocus()
    expect((await axe.run(container, { rules: { 'color-contrast': { enabled: false } } })).violations).toEqual([])

    await user.click(screen.getByRole('tab', { name: 'Ejercicios' }))
    await user.click(await screen.findByRole('button', { name: 'Editar Sentadilla' }))
    expect(screen.getByRole('heading', { name: 'Editar ejercicio' })).toHaveFocus()
    expect((await axe.run(container, { rules: { 'color-contrast': { enabled: false } } })).violations).toEqual([])
  })
})

function browserToken(expiresAt: number): string {
  const payload = btoa(JSON.stringify({ exp: Math.floor(expiresAt / 1000) }))
  return `header.${payload}.signature`
}
