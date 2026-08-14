import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { ExercisesView } from "./ExercisesView";

const response = (body: unknown, status = 200) =>
  Promise.resolve(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    }),
  );

describe("ExercisesView FitPro", () => {
  beforeEach(() => {
    sessionStorage.setItem("accessToken", "token");
    vi.restoreAllMocks();
  });

  it("creates through freely navigable wizard without follow-up GET", async () => {
    const fetchMock = vi.fn((_input: RequestInfo | URL, init?: RequestInit) =>
      init?.method === "POST"
        ? response({ id: 2, name: "Plancha" }, 201)
        : response([]),
    );
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<ExercisesView onUnauthenticated={vi.fn()} />);
    await screen.findByText("Todavía no creaste ejercicios.");
    await user.click(screen.getByRole("button", { name: "Crear ejercicio" }));
    const nameField = screen.getByLabelText("Nombre");
    expect(nameField.closest("label")).toHaveClass("grid");
    await user.type(nameField, "Plancha");
    await user.type(
      screen.getByLabelText("URL de imagen (opcional)"),
      "https://example.test/plancha.webp",
    );
    await user.type(
      screen.getByLabelText("URL de video (opcional)"),
      "https://youtu.be/dQw4w9WgXcQ",
    );
    await user.click(screen.getByRole("tab", { name: /Resumen/ }));
    const imageRegion = screen.getByRole("region", { name: "Imagen" });
    expect(
      within(imageRegion).getByAltText("Demostración de Plancha"),
    ).toBeInTheDocument();
    expect(imageRegion.parentElement).toHaveClass("md:grid-cols-2");
    expect(
      within(screen.getByRole("region", { name: "Video" })).getByTitle(
        "Video de ejecución de Plancha",
      ),
    ).toBeInTheDocument();
    await user.click(
      within(screen.getByRole("dialog")).getByRole("button", {
        name: "Crear ejercicio",
      }),
    );
    expect(
      await screen.findByRole("heading", { name: "Plancha" }),
    ).toBeInTheDocument();
    expect(
      fetchMock.mock.calls.filter(([, init]) => init?.method === "POST"),
    ).toHaveLength(1);
    expect(fetchMock.mock.calls).toHaveLength(2);
  });

  it("protects dirty dismissal and exposes independent detail cards", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        response([
          { id: 3, name: "Remo" },
          { id: 4, name: "Plancha" },
        ]),
      ),
    );
    const user = userEvent.setup();
    render(<ExercisesView onUnauthenticated={vi.fn()} />);
    await screen.findByText("Remo");
    await user.click(screen.getByRole("button", { name: "Editar Remo" }));
    await user.clear(screen.getByLabelText("Nombre"));
    await user.type(screen.getByLabelText("Nombre"), "Remo nuevo");
    await user.click(screen.getByRole("button", { name: "Cancelar" }));
    expect(
      screen.getByRole("alertdialog", { name: "¿Salir del wizard?" }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Nombre")).toHaveValue("Remo nuevo");
    await user.click(screen.getByRole("button", { name: "Seguir editando" }));
    expect(
      screen.queryByRole("alertdialog", { name: "¿Salir del wizard?" }),
    ).not.toBeInTheDocument();
  });

  it("expands an exercise card without opening an overlay", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        response([
          {
            id: 3,
            name: "Remo",
            description: "Trabajo de espalda",
            imageUrl: "https://example.test/remo.webp",
            videoUrl: "https://youtu.be/dQw4w9WgXcQ",
          },
        ]),
      ),
    );
    const user = userEvent.setup();
    render(<ExercisesView onUnauthenticated={vi.fn()} />);

    const toggle = await screen.findByRole("button", {
      name: "Ver detalles de Remo",
    });
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    const image = screen.getByAltText("Demostración de Remo");
    expect(image).toBeInTheDocument();
    expect(
      screen.queryByTitle("Video de ejecución de Remo"),
    ).not.toBeInTheDocument();

    await user.click(image);

    const collapseToggle = screen.getByRole("button", {
      name: "Ocultar detalles de Remo",
    });
    expect(collapseToggle).toHaveAttribute("aria-expanded", "true");
    expect(
      screen.queryByAltText("Demostración de Remo"),
    ).not.toBeInTheDocument();
    const video = screen.getByTitle("Video de ejecución de Remo");
    const description = screen.getByText("Trabajo de espalda");
    expect(video).toBeInTheDocument();
    expect(description).toBeVisible();
    expect(
      video.compareDocumentPosition(description) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(document.querySelector("dialog[open]")).not.toBeInTheDocument();
  });
});
