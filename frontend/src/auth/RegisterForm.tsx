import { type FormEvent, useEffect, useRef, useState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { register } from "./api";
import { ApiError, type RegisteredUser, type RegisterInput } from "./types";

interface RegisterFormProps {
  onRegistered: (user: RegisteredUser) => void;
  onShowLogin?: () => void;
}

const emptyForm: RegisterInput = {
  firstName: "",
  lastName: "",
  phone: "",
  address: { street: "", number: "", apartment: "", city: "", province: "" },
  username: "",
  email: "",
  password: "",
};

export function RegisterForm({ onRegistered, onShowLogin }: RegisterFormProps) {
  const [form, setForm] = useState<RegisterInput>(emptyForm);
  const [passwordConfirmation, setPasswordConfirmation] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showPasswordConfirmation, setShowPasswordConfirmation] =
    useState(false);
  const [pending, setPending] = useState(false);
  const [errors, setErrors] = useState<Record<string, string[]>>({});
  const [created, setCreated] = useState<RegisteredUser | null>(null);
  const [generalError, setGeneralError] = useState("");
  const summaryRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (generalError || Object.keys(errors).length > 0) {
      summaryRef.current?.focus();
    }
  }, [errors, generalError]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrors({});
    setGeneralError("");
    setCreated(null);

    if (form.password !== passwordConfirmation) {
      setErrors({ passwordConfirmation: ["Las contraseñas deben coincidir."] });
      setGeneralError("Revisá los campos indicados.");
      return;
    }

    setPending(true);
    try {
      const user = await register(form);
      setCreated(user);
      clearPasswords();
      onRegistered(user);
    } catch (error) {
      clearPasswords();
      if (error instanceof ApiError) {
        setErrors(error.body.fields ?? {});
        setGeneralError(error.message);
      } else {
        setGeneralError("No se pudo conectar con el servidor.");
      }
    } finally {
      setPending(false);
    }
  }

  function clearPasswords() {
    setForm((current) => ({ ...current, password: "" }));
    setPasswordConfirmation("");
    setShowPassword(false);
    setShowPasswordConfirmation(false);
  }

  function updateField(
    field: keyof Omit<RegisterInput, "address">,
    value: string,
  ) {
    setForm((current) => ({ ...current, [field]: value }));
    if (field === "password") clearPasswordMismatch();
  }

  function updatePasswordConfirmation(value: string) {
    setPasswordConfirmation(value);
    clearPasswordMismatch();
  }

  function clearPasswordMismatch() {
    setErrors((current) => {
      if (!current.passwordConfirmation) return current;
      const remaining = { ...current };
      delete remaining.passwordConfirmation;
      return remaining;
    });
    setGeneralError((current) =>
      current === "Revisá los campos indicados." ? "" : current,
    );
  }

  function updateAddress(field: keyof RegisterInput["address"], value: string) {
    setForm((current) => ({
      ...current,
      address: { ...current.address, [field]: value },
    }));
  }

  return (
    <form aria-label="Registro" className="card-body gap-4" onSubmit={submit}>
      <fieldset
        className="fieldset grid gap-4 md:grid-cols-2"
        disabled={pending}
      >
        <legend className="fieldset-legend col-span-full text-lg">
          Datos personales
        </legend>
        <TextField
          autoComplete="given-name"
          label="Nombre"
          value={form.firstName}
          onChange={(value) => updateField("firstName", value)}
          errors={errors.firstName}
        />
        <TextField
          autoComplete="family-name"
          label="Apellido"
          value={form.lastName}
          onChange={(value) => updateField("lastName", value)}
          errors={errors.lastName}
        />
        <TextField
          autoComplete="tel"
          label="Teléfono"
          type="tel"
          value={form.phone}
          onChange={(value) => updateField("phone", value)}
          errors={errors.phone}
        />
        <TextField
          autoComplete="address-line1"
          label="Calle"
          value={form.address.street}
          onChange={(value) => updateAddress("street", value)}
          errors={errors.street}
        />
        <TextField
          autoComplete="address-line2"
          label="Número"
          value={form.address.number}
          onChange={(value) => updateAddress("number", value)}
          errors={errors.number}
        />
        <TextField
          autoComplete="address-line3"
          label="Departamento (opcional)"
          required={false}
          value={form.address.apartment}
          onChange={(value) => updateAddress("apartment", value)}
          errors={errors.apartment}
        />
        <TextField
          autoComplete="address-level2"
          label="Ciudad"
          value={form.address.city}
          onChange={(value) => updateAddress("city", value)}
          errors={errors.city}
        />
        <TextField
          autoComplete="address-level1"
          label="Provincia"
          value={form.address.province}
          onChange={(value) => updateAddress("province", value)}
          errors={errors.province}
        />
        <TextField
          autoComplete="username"
          label="Nombre de usuario"
          value={form.username}
          onChange={(value) => updateField("username", value)}
          errors={errors.username}
        />
        <TextField
          autoComplete="email"
          label="Email"
          type="email"
          value={form.email}
          onChange={(value) => updateField("email", value)}
          errors={errors.email}
        />
        <PasswordField
          id="password"
          label="Contraseña"
          toggleContext="contraseña ingresada"
          value={form.password}
          visible={showPassword}
          onChange={(value) => updateField("password", value)}
          onToggle={() => setShowPassword((current) => !current)}
          errors={errors.password}
        />
        <PasswordField
          id="password-confirmation"
          label="Confirmar contraseña"
          toggleContext="contraseña de confirmación"
          value={passwordConfirmation}
          visible={showPasswordConfirmation}
          onChange={updatePasswordConfirmation}
          onToggle={() => setShowPasswordConfirmation((current) => !current)}
          errors={errors.passwordConfirmation}
        />
      </fieldset>

      {(generalError || Object.keys(errors).length > 0) && (
        <div
          ref={summaryRef}
          className="alert alert-error"
          role="alert"
          tabIndex={-1}
        >
          <span>{generalError || "Revisá los campos indicados."}</span>
        </div>
      )}
      {created && (
        <div className="alert alert-success" role="status">
          <span>Cuenta creada</span>
          <span>{created.username}</span>
          <span>{created.email}</span>
        </div>
      )}

      <button className="btn btn-primary" disabled={pending} type="submit">
        {pending && (
          <span className="loading loading-spinner" aria-hidden="true" />
        )}
        {pending ? "Creando cuenta" : "Crear cuenta"}
      </button>
      {onShowLogin && (
        <button
          type="button"
          className="link link-secondary"
          onClick={onShowLogin}
        >
          ¿Ya tienes cuenta? Inicia sesión
        </button>
      )}
    </form>
  );
}

interface PasswordFieldProps {
  id: string;
  label: string;
  toggleContext: string;
  value: string;
  visible: boolean;
  onChange: (value: string) => void;
  onToggle: () => void;
  errors?: string[];
}

function PasswordField({
  id,
  label,
  toggleContext,
  value,
  visible,
  onChange,
  onToggle,
  errors = [],
}: PasswordFieldProps) {
  const errorID = `${id}-error`;
  const action = visible ? "Ocultar" : "Ver";
  // El nombre contextual diferencia ambos controles y conserva el texto visible.
  // aria-pressed comunica el estado sin crear un widget personalizado.
  const toggleName = `${action} ${toggleContext}`;

  return (
    <div className="fieldset-label flex flex-col items-stretch gap-1">
      <label htmlFor={id}>{label}</label>
      <div className="flex flex-col gap-2 sm:flex-row">
        <input
          aria-describedby={errors.length > 0 ? errorID : undefined}
          aria-invalid={errors.length > 0}
          autoComplete="new-password"
          className={`input min-w-0 flex-1 ${errors.length > 0 ? "input-error" : ""}`}
          id={id}
          onChange={(event) => onChange(event.target.value)}
          required
          type={visible ? "text" : "password"}
          value={value}
        />
        <button
          aria-controls={id}
          aria-label={toggleName}
          aria-pressed={visible}
          className="btn btn-outline min-h-11 shrink-0"
          onClick={onToggle}
          type="button"
        >
          {visible ? (
            <EyeOff aria-hidden="true" size={18} />
          ) : (
            <Eye aria-hidden="true" size={18} />
          )}
          {action} contraseña
        </button>
      </div>
      {errors.length > 0 && (
        <span className="text-error text-sm" id={errorID}>
          {errors.join(" ")}
        </span>
      )}
    </div>
  );
}

interface TextFieldProps {
  autoComplete: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  errors?: string[];
  required?: boolean;
  type?: string;
}

function TextField({
  autoComplete,
  label,
  value,
  onChange,
  errors = [],
  required = true,
  type = "text",
}: TextFieldProps) {
  const id = label.toLowerCase().replaceAll(/[^a-z0-9]+/g, "-");
  const errorID = `${id}-error`;
  return (
    <label
      className="fieldset-label flex flex-col items-stretch gap-1"
      htmlFor={id}
    >
      <span>{label}</span>
      <input
        aria-label={label}
        aria-describedby={errors.length > 0 ? errorID : undefined}
        aria-invalid={errors.length > 0}
        className={`input w-full ${errors.length > 0 ? "input-error" : ""}`}
        autoComplete={autoComplete}
        id={id}
        onChange={(event) => onChange(event.target.value)}
        required={required}
        type={type}
        value={value}
      />
      {errors.length > 0 && (
        <span className="text-error text-sm" id={errorID}>
          {errors.join(" ")}
        </span>
      )}
    </label>
  );
}
