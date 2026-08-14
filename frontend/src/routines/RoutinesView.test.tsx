import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { RoutinesView } from "./RoutinesView";

const response = (body: unknown, status = 200) =>
  Promise.resolve(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    }),
  );

it("adds the same session repeatedly on different days", async () => {
  sessionStorage.setItem("accessToken", "token");
  const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
    if (init?.method === "POST")
      return response({ id: 7, name: "Semana", sessions: [] }, 201);
    if (String(input) === "/api/sessions")
      return response([{ id: 3, name: "Piernas", exerciseCount: 2 }]);
    return response([]);
  });
  vi.stubGlobal("fetch", fetchMock);
  const user = userEvent.setup();
  render(<RoutinesView onUnauthenticated={vi.fn()} />);
  await user.click(await screen.findByRole("button", { name: "Crear rutina" }));
  await user.type(screen.getByLabelText("Nombre"), "Semana");
  await user.click(screen.getByRole("tab", { name: /Sesiones/ }));
  const add = screen.getByRole("button", { name: /Piernas.*Agregar/ });
  await user.click(add);
  await user.click(add);
  const selects = screen.getAllByRole("combobox");
  await user.selectOptions(selects[0], "1");
  await user.selectOptions(selects[1], "4");
  await user.click(screen.getByRole("tab", { name: /Resumen/ }));
  await user.click(
    within(screen.getByRole("dialog")).getByRole("button", {
      name: "Crear rutina",
    }),
  );
  const call = fetchMock.mock.calls.find(([, init]) => init?.method === "POST");
  expect(JSON.parse(String(call?.[1]?.body)).sessions).toEqual([
    { sessionId: 3, day: 1 },
    { sessionId: 3, day: 4 },
  ]);
});
