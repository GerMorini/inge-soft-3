import { fireEvent, render, screen } from "@testing-library/react";
import { AuthShell } from "./AuthShell";

it("keeps authentication usable when decorative image fails", () => {
  const { container } = render(
    <AuthShell title="Iniciar sesión">
      <button>Ingresar</button>
    </AuthShell>,
  );
  const image = container.querySelector("img");
  expect(image).toHaveAttribute("alt", "");
  fireEvent.error(image!);
  expect(container.querySelector("img")).toBeNull();
  expect(screen.getByRole("button", { name: "Ingresar" })).toBeInTheDocument();
});
