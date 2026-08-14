import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import App from "./App";

const response = (body: unknown) =>
  Promise.resolve(
    new Response(JSON.stringify(body), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    }),
  );
const token = () =>
  `header.${btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 60 }))}.signature`;

describe("FitPro shell", () => {
  beforeEach(() => {
    sessionStorage.clear();
    vi.restoreAllMocks();
  });
  it("starts at login and switches to complete registration without navbar", async () => {
    const user = userEvent.setup();
    render(<App />);
    expect(
      screen.getByRole("heading", { name: "Iniciar sesión" }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
    await user.click(
      screen.getByRole("button", {
        name: "¿No tienes cuenta? Créate una aquí",
      }),
    );
    expect(screen.getByRole("form", { name: "Registro" })).toBeInTheDocument();
    expect(screen.getByLabelText("Confirmar contraseña")).toBeInTheDocument();
  });
  it("renders keyboard navigation and heroes after authentication", async () => {
    sessionStorage.setItem("accessToken", token());
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL) =>
        String(input) === "/api/auth/me"
          ? response({ id: 1, username: "ada" })
          : response([]),
      ),
    );
    const user = userEvent.setup();
    render(<App />);
    expect(await screen.findByText("FitPro")).toBeInTheDocument();
    const navigation = screen.getByRole("navigation", {
      name: "Navegación principal",
    });
    expect(
      within(navigation).getByRole("group", {
        name: "Controles de sesión",
      }),
    ).toHaveTextContent("ada");
    expect(
      within(navigation).getByRole("button", {
        name: "Cerrar sesión de ada",
      }),
    ).toBeInTheDocument();
    expect(within(navigation).getByRole("tablist")).toHaveClass(
      "justify-self-center",
    );
    const routines = screen.getByRole("tab", { name: "Rutinas" });
    routines.focus();
    await user.keyboard("{ArrowRight}");
    expect(screen.getByRole("tab", { name: "Sesiones" })).toHaveFocus();
    expect(
      await screen.findByRole("heading", { name: "Sesiones" }),
    ).toBeInTheDocument();
  });
});
