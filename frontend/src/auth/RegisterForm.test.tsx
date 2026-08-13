import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { RegisterForm } from './RegisterForm'

const validUser = {
  id: 1,
  username: 'ada_01',
  email: 'ada@example.com',
}

describe('RegisterForm', () => {
  it('shows every field and registers valid data without apartment', async () => {
    const user = userEvent.setup()
    const onRegistered = vi.fn()
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(validUser), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    render(<RegisterForm onRegistered={onRegistered} />)

    await completeRegistrationForm(user)
    expect(screen.getByLabelText('Departamento (opcional)')).not.toBeRequired()
    await user.click(screen.getByRole('button', { name: 'Crear cuenta' }))

    expect(await screen.findByText('Cuenta creada')).toBeInTheDocument()
    expect(screen.getByText('ada_01')).toBeInTheDocument()
    expect(screen.getByText('ada@example.com')).toBeInTheDocument()
    expect(onRegistered).toHaveBeenCalledWith(validUser)
    expect(screen.getByLabelText('Contraseña')).toHaveValue('')
    expect(screen.getByLabelText('Confirmar contraseña')).toHaveValue('')
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).not.toHaveProperty(
      'passwordConfirmation',
    )
  })

  it('disables submission while request is pending', async () => {
    const user = userEvent.setup()
    vi.spyOn(globalThis, 'fetch').mockImplementation(() => new Promise(() => {}))
    render(<RegisterForm onRegistered={() => {}} />)

    await completeRegistrationForm(user)
    await user.click(screen.getByRole('button', { name: 'Crear cuenta' }))

    expect(screen.getByRole('button', { name: 'Creando cuenta' })).toBeDisabled()
  })

  it('shows field errors, preserves safe values and clears password', async () => {
    const user = userEvent.setup()
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({
        error: {
          code: 'validation_failed',
          message: 'Revisá los campos indicados.',
          fields: { username: ['No admite espacios.'], password: ['Debe incluir 3 números.'] },
        },
      }), { status: 400, headers: { 'Content-Type': 'application/json' } }),
    )
    render(<RegisterForm onRegistered={() => {}} />)

    await completeRegistrationForm(user)
    await user.click(screen.getByRole('button', { name: 'Crear cuenta' }))

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveFocus()
    expect(screen.getByText('No admite espacios.')).toBeInTheDocument()
    expect(screen.getByLabelText('Nombre')).toHaveValue('Ada')
    expect(screen.getByLabelText('Contraseña')).toHaveValue('')
    expect(screen.getByLabelText('Confirmar contraseña')).toHaveValue('')
  })

  it('blocks mismatched passwords before calling the API', async () => {
    const user = userEvent.setup()
    const fetchMock = vi.spyOn(globalThis, 'fetch')
    render(<RegisterForm onRegistered={() => {}} />)

    await completeRegistrationForm(user, 'Segura!@123', 'Segura!@124')
    await user.click(screen.getByRole('button', { name: 'Crear cuenta' }))

    expect(fetchMock).not.toHaveBeenCalled()
    expect(await screen.findByText('Las contraseñas deben coincidir.')).toBeInTheDocument()
    expect(screen.getByLabelText('Confirmar contraseña')).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByRole('alert')).toHaveFocus()
    expect(screen.getByLabelText('Contraseña')).toHaveValue('Segura!@123')
    expect(screen.getByLabelText('Confirmar contraseña')).toHaveValue('Segura!@124')

    await user.clear(screen.getByLabelText('Confirmar contraseña'))
    await user.type(screen.getByLabelText('Confirmar contraseña'), 'Segura!@123')
    expect(screen.queryByText('Las contraseñas deben coincidir.')).not.toBeInTheDocument()
    expect(screen.getByLabelText('Confirmar contraseña')).toHaveAttribute('aria-invalid', 'false')
  })

  it('shows and hides both passwords independently with keyboard controls', async () => {
    const user = userEvent.setup()
    render(<RegisterForm onRegistered={() => {}} />)

    const password = screen.getByLabelText('Contraseña')
    const confirmation = screen.getByLabelText('Confirmar contraseña')
    await user.type(password, 'Segura!@123')
    await user.type(confirmation, 'Segura!@123')

    expect(password).toHaveAttribute('type', 'password')
    expect(confirmation).toHaveAttribute('type', 'password')
    expect(screen.getAllByText('Ver contraseña')).toHaveLength(2)

    password.focus()
    await user.tab()
    const passwordToggle = screen.getByRole('button', {
      name: 'Ver contraseña ingresada',
    })
    expect(passwordToggle).toHaveFocus()
    await user.keyboard('{Enter}')

    expect(password).toHaveAttribute('type', 'text')
    expect(confirmation).toHaveAttribute('type', 'password')
    expect(passwordToggle).toHaveAttribute('aria-pressed', 'true')
    expect(passwordToggle).toHaveAccessibleName('Ocultar contraseña ingresada')

    await user.click(screen.getByRole('button', {
      name: 'Ver contraseña de confirmación',
    }))
    expect(password).toHaveAttribute('type', 'text')
    expect(confirmation).toHaveAttribute('type', 'text')
    expect(password).toHaveValue('Segura!@123')
    expect(confirmation).toHaveValue('Segura!@123')

    passwordToggle.focus()
    await user.keyboard(' ')
    expect(password).toHaveAttribute('type', 'password')
    expect(confirmation).toHaveAttribute('type', 'text')
  })
})

async function completeRegistrationForm(
  user: ReturnType<typeof userEvent.setup>,
  password = 'Segura!@123',
  confirmation = password,
) {
  await user.type(screen.getByLabelText('Nombre'), 'Ada')
  await user.type(screen.getByLabelText('Apellido'), 'Lovelace')
  await user.type(screen.getByLabelText('Teléfono'), '+5493515551234')
  await user.type(screen.getByLabelText('Calle'), 'San Martín')
  await user.type(screen.getByLabelText('Número'), '123 Bis')
  await user.type(screen.getByLabelText('Ciudad'), 'Córdoba')
  await user.type(screen.getByLabelText('Provincia'), 'Córdoba')
  await user.type(screen.getByLabelText('Nombre de usuario'), 'Ada_01')
  await user.type(screen.getByLabelText('Email'), 'ADA@example.com')
  await user.type(screen.getByLabelText('Contraseña'), password)
  await user.type(screen.getByLabelText('Confirmar contraseña'), confirmation)
}
