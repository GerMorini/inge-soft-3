import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { vi } from "vitest";
import { ModalDialog } from "./ModalDialog";

it("opens accessibly and closes from backdrop", async () => {
  const closed = vi.fn();
  function Harness() {
    const [open, setOpen] = useState(false);
    return (
      <>
        <button onClick={() => setOpen(true)}>Abrir</button>
        <ModalDialog
          open={open}
          title="Detalle"
          onRequestClose={(reason) => {
            closed(reason);
            setOpen(false);
          }}
        >
          Contenido
        </ModalDialog>
      </>
    );
  }
  const user = userEvent.setup();
  render(<Harness />);
  const trigger = screen.getByRole("button", { name: "Abrir" });
  await user.click(trigger);
  expect(screen.getByRole("dialog", { name: "Detalle" })).toBeInTheDocument();
  await user.click(
    screen.getAllByRole("button", { name: "Cerrar Detalle" })[1],
  );
  expect(closed).toHaveBeenCalledWith("backdrop");
});
