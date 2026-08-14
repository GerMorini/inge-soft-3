import { type FormEvent, useState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { login } from "./api";
import { storeAccessToken } from "./session";
import { ApiError } from "./types";

interface LoginFormProps {
  onAuthenticated: () => void;
  onShowRegister?: () => void;
}

export function LoginForm({ onAuthenticated, onShowRegister }: LoginFormProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState("");
  const [success, setSuccess] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setMessage("");
    setSuccess(false);
    try {
      const result = await login({ username, password });
      storeAccessToken(result.accessToken);
      setPassword("");
      setSuccess(true);
      onAuthenticated();
    } catch (error) {
      setPassword("");
      setMessage(
        error instanceof ApiError
          ? error.message
          : "No se pudo conectar con el servidor.",
      );
    } finally {
      setPending(false);
    }
  }

  return (
    <form
      aria-label="Inicio de sesión"
      className="card-body gap-4"
      onSubmit={submit}
    >
      <label
        className="fieldset-label flex flex-col items-stretch gap-1"
        htmlFor="login-username"
      >
        <span>Nombre de usuario</span>
        <input
          autoComplete="username"
          className="input w-full"
          id="login-username"
          onChange={(event) => setUsername(event.target.value)}
          required
          value={username}
        />
      </label>
      <label
        className="fieldset-label flex flex-col items-stretch gap-1"
        htmlFor="login-password"
      >
        <span>Contraseña</span>
        <span className="join w-full">
          <input
            autoComplete="current-password"
            className="input join-item min-w-0 flex-1"
            id="login-password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type={showPassword ? "text" : "password"}
            value={password}
          />
          <button
            type="button"
            className="btn join-item min-h-11"
            aria-label={showPassword ? "Ocultar contraseña" : "Ver contraseña"}
            aria-pressed={showPassword}
            onClick={() => setShowPassword((current) => !current)}
          >
            {showPassword ? (
              <EyeOff aria-hidden="true" size={18} />
            ) : (
              <Eye aria-hidden="true" size={18} />
            )}
          </button>
        </span>
      </label>
      {message && (
        <div className="alert alert-error" role="alert">
          {message}
        </div>
      )}
      {success && (
        <div className="alert alert-success" role="status">
          Sesión iniciada
        </div>
      )}
      <button className="btn btn-primary" disabled={pending} type="submit">
        {pending && (
          <span className="loading loading-spinner" aria-hidden="true" />
        )}
        {pending ? "Ingresando" : "Ingresar"}
      </button>
      {onShowRegister && (
        <button
          type="button"
          className="link link-secondary"
          onClick={onShowRegister}
        >
          ¿No tienes cuenta? Créate una aquí
        </button>
      )}
    </form>
  );
}
