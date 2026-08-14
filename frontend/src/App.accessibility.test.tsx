import axe from "axe-core";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import App from "./App";

it("has no automated structural violations in auth modes", async () => {
  sessionStorage.clear();
  const user = userEvent.setup();
  const { container } = render(<App />);
  expect(
    (
      await axe.run(container, {
        rules: { "color-contrast": { enabled: false } },
      })
    ).violations,
  ).toEqual([]);
  await user.click(
    screen.getByRole("button", { name: "¿No tienes cuenta? Créate una aquí" }),
  );
  await user.click(
    screen.getByRole("button", { name: "Ver contraseña ingresada" }),
  );
  await user.click(
    screen.getByRole("button", { name: "Ver contraseña de confirmación" }),
  );
  expect(
    (
      await axe.run(container, {
        rules: { "color-contrast": { enabled: false } },
      })
    ).violations,
  ).toEqual([]);
});

it("has no automated structural violations in workspace", async () => {
  const payload = btoa(
    JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 60 }),
  );
  sessionStorage.setItem("accessToken", `h.${payload}.s`);
  vi.stubGlobal(
    "fetch",
    vi.fn((input: RequestInfo | URL) =>
      Promise.resolve(
        new Response(
          JSON.stringify(
            String(input) === "/api/auth/me" ? { id: 1, username: "ada" } : [],
          ),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      ),
    ),
  );
  const { container } = render(<App />);
  await screen.findByText("FitPro");
  expect(
    (
      await axe.run(container, {
        rules: { "color-contrast": { enabled: false } },
      })
    ).violations,
  ).toEqual([]);
});
