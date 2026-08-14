import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { SessionsView } from "./SessionsView";

const response = (body: unknown, status = 200) =>
  Promise.resolve(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    }),
  );

it("uses backend count and derives count after save", async () => {
  sessionStorage.setItem("accessToken", "token");
  const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
    if (init?.method === "POST")
      return response(
        {
          id: 8,
          name: "Piernas",
          exercises: [
            {
              exercise: { id: 1, name: "Sentadilla" },
              series: 3,
              repetitions: 10,
              order: 1,
            },
          ],
        },
        201,
      );
    if (String(input) === "/api/sessions")
      return response([{ id: 2, name: "Vacía", exerciseCount: 0 }]);
    return response([{ id: 1, name: "Sentadilla" }]);
  });
  vi.stubGlobal("fetch", fetchMock);
  const user = userEvent.setup();
  render(<SessionsView onUnauthenticated={vi.fn()} />);
  expect(await screen.findByText(/0 ejercicios/)).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Crear sesión" }));
  await user.type(screen.getByLabelText("Nombre"), "Piernas");
  await user.click(screen.getByRole("tab", { name: /Ejercicios/ }));
  await user.click(screen.getByRole("checkbox", { name: "Sentadilla" }));
  await user.click(screen.getByRole("tab", { name: /Resumen/ }));
  await user.click(
    within(screen.getByRole("dialog")).getByRole("button", {
      name: "Crear sesión",
    }),
  );
  expect(
    await screen.findByRole("heading", { name: "Piernas" }),
  ).toBeInTheDocument();
  expect(screen.getAllByText(/1 ejercicios/).length).toBeGreaterThan(0);
  expect(
    fetchMock.mock.calls.filter(([, init]) => init?.method === "POST"),
  ).toHaveLength(1);
});
