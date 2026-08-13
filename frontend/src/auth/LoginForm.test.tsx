import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { LoginForm } from './LoginForm'
import { accessTokenKey } from './session'

describe('LoginForm', () => {
  beforeEach(() => sessionStorage.clear())

  it('stores successful login token and reports identity transition', async () => {
    const user = userEvent.setup()
    const onAuthenticated = vi.fn()
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ accessToken: 'header.payload.signature', tokenType: 'Bearer', expiresIn: 1800 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    render(<LoginForm onAuthenticated={onAuthenticated} />)

    await user.type(screen.getByLabelText('Nombre de usuario'), 'Ada_01')
    await user.type(screen.getByLabelText('Contraseña'), 'Segura!@123')
    await user.click(screen.getByRole('button', { name: 'Ingresar' }))

    expect(await screen.findByText('Sesión iniciada')).toBeInTheDocument()
    expect(sessionStorage.getItem(accessTokenKey)).toBe('header.payload.signature')
    expect(onAuthenticated).toHaveBeenCalled()
    expect(screen.getByLabelText('Contraseña')).toHaveValue('')
  })

  it('shows the same generic credential failure', async () => {
    const user = userEvent.setup()
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: { code: 'invalid_credentials', message: 'Username o contraseña inválidos.' } }), {
        status: 401,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    render(<LoginForm onAuthenticated={() => {}} />)

    await user.type(screen.getByLabelText('Nombre de usuario'), 'nadie')
    await user.type(screen.getByLabelText('Contraseña'), 'Incorrecta!123')
    await user.click(screen.getByRole('button', { name: 'Ingresar' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Username o contraseña inválidos.')
    expect(sessionStorage.getItem(accessTokenKey)).toBeNull()
  })
})
