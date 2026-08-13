import axe from 'axe-core'
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'

describe('App accessibility', () => {
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
})
