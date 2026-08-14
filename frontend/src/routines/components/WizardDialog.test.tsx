import { useState } from "react";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { WizardDialog } from "./WizardDialog";

it("navigates freely and only saves from summary", async () => {
  const save = vi.fn();
  const change = vi.fn();
  const user = userEvent.setup();
  const { rerender } = render(
    <WizardDialog
      open
      title="Crear"
      activeStep={0}
      dirty={false}
      saving={false}
      saveLabel="Guardar"
      onStepChange={change}
      onClose={vi.fn()}
      onSubmit={save}
      steps={[
        { label: "Datos", content: <input aria-label="Dato" /> },
        { label: "Resumen", content: "Resumen" },
      ]}
    />,
  );
  expect(
    within(screen.getByRole("dialog")).queryByRole("button", {
      name: "Guardar",
    }),
  ).not.toBeInTheDocument();
  await user.click(screen.getByRole("tab", { name: /Resumen/ }));
  expect(change).toHaveBeenCalledWith(1);
  rerender(
    <WizardDialog
      open
      title="Crear"
      activeStep={1}
      dirty={false}
      saving={false}
      saveLabel="Guardar"
      onStepChange={change}
      onClose={vi.fn()}
      onSubmit={save}
      steps={[
        { label: "Datos", content: "Datos" },
        { label: "Resumen", content: "Resumen" },
      ]}
    />,
  );
  await user.click(
    within(screen.getByRole("dialog")).getByRole("button", { name: "Guardar" }),
  );
  expect(save).toHaveBeenCalledOnce();
});

it("shows summary without saving from the penultimate step", async () => {
  const save = vi.fn();
  const user = userEvent.setup();

  function Harness() {
    const [activeStep, setActiveStep] = useState(1);
    return (
      <WizardDialog
        open
        title="Crear sesión"
        activeStep={activeStep}
        dirty
        saving={false}
        saveLabel="Crear sesión"
        onStepChange={setActiveStep}
        onClose={vi.fn()}
        onSubmit={save}
        steps={[
          { label: "Datos básicos", content: "Datos" },
          { label: "Ejercicios", content: "Selección" },
          { label: "Resumen", content: "Vista resumida" },
        ]}
      />
    );
  }

  render(<Harness />);
  await user.click(screen.getByRole("button", { name: /Siguiente/ }));

  expect(save).not.toHaveBeenCalled();
  expect(screen.getByText("Vista resumida")).toBeVisible();
  expect(
    screen.getByRole("button", { name: "Crear sesión" }),
  ).toBeInTheDocument();
});

it("uses a custom confirmation before discarding a dirty wizard", async () => {
  const close = vi.fn();
  const user = userEvent.setup();

  render(
    <WizardDialog
      open
      title="Editar ejercicio"
      activeStep={0}
      dirty
      saving={false}
      saveLabel="Guardar"
      onStepChange={vi.fn()}
      onClose={close}
      onSubmit={vi.fn()}
      steps={[{ label: "Datos", content: "Formulario" }]}
    />,
  );

  await user.click(screen.getByRole("button", { name: "Cancelar" }));
  expect(close).not.toHaveBeenCalled();
  expect(
    screen.getByRole("alertdialog", { name: "¿Salir del wizard?" }),
  ).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Seguir editando" }));
  expect(close).not.toHaveBeenCalled();

  await user.click(screen.getByRole("button", { name: "Cancelar" }));
  await user.click(screen.getByRole("button", { name: "Salir sin guardar" }));
  expect(close).toHaveBeenCalledOnce();
});
