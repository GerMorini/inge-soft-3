import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { SessionStatus } from "./SessionStatus";
import { accessTokenKey } from "./session";

describe("SessionStatus", () => {
  beforeEach(() => sessionStorage.clear());

  it("restores current user with a bearer token", async () => {
    const token = browserToken(Date.now() + 60_000);
    sessionStorage.setItem(accessTokenKey, token);
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 42, username: "ada_01" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    render(
      <SessionStatus onUnauthenticated={() => {}}>
        {({ currentUser }) => <span>{currentUser.username}</span>}
      </SessionStatus>,
    );

    expect(await screen.findByText("ada_01")).toBeInTheDocument();
    expect(globalThis.fetch).toHaveBeenCalledWith("/api/auth/me", {
      headers: { Authorization: `Bearer ${token}` },
    });
  });

  it("clears rejected tokens and returns to unauthenticated state", async () => {
    sessionStorage.setItem(accessTokenKey, "expired-token");
    const onUnauthenticated = vi.fn();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          error: {
            code: "invalid_token",
            message: "Token de acceso inválido o vencido.",
          },
        }),
        {
          status: 401,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );
    render(<SessionStatus onUnauthenticated={onUnauthenticated} />);

    await waitFor(() => expect(onUnauthenticated).toHaveBeenCalled());
    expect(sessionStorage.getItem(accessTokenKey)).toBeNull();
  });

  it("logs out by removing frontend token", async () => {
    const user = userEvent.setup();
    sessionStorage.setItem(accessTokenKey, browserToken(Date.now() + 60_000));
    const onUnauthenticated = vi.fn();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: 42, username: "ada_01" }), {
        status: 200,
      }),
    );
    render(
      <SessionStatus onUnauthenticated={onUnauthenticated}>
        {({ currentUser, logout }) => (
          <button type="button" onClick={logout}>
            Cerrar sesión de {currentUser.username}
          </button>
        )}
      </SessionStatus>,
    );
    await user.click(
      await screen.findByRole("button", { name: "Cerrar sesión de ada_01" }),
    );
    expect(sessionStorage.getItem(accessTokenKey)).toBeNull();
    expect(onUnauthenticated).toHaveBeenCalled();
  });
});

function browserToken(expiresAt: number): string {
  const payload = btoa(JSON.stringify({ exp: Math.floor(expiresAt / 1000) }));
  return `header.${payload}.signature`;
}
