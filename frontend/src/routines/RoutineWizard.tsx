import { useEffect, useMemo, useRef, useState } from "react";
import { Plus, X } from "lucide-react";
import type { RoutineDetail, RoutineInput, SessionSummary } from "./types";
import { SearchableCatalog } from "./components/SearchableCatalog";
import { WizardDialog } from "./components/WizardDialog";

const days = [
  "Lunes",
  "Martes",
  "Miércoles",
  "Jueves",
  "Viernes",
  "Sábado",
  "Domingo",
];
type Row = { key: number; sessionId: number; day: number };
interface Props {
  open: boolean;
  initial?: RoutineDetail;
  sessions: SessionSummary[];
  saving: boolean;
  onClose: () => void;
  onDirtyChange: (dirty: boolean) => void;
  onSave: (input: RoutineInput) => Promise<void>;
}

export function RoutineWizard({
  open,
  initial,
  sessions,
  saving,
  onClose,
  onDirtyChange,
  onSave,
}: Props) {
  const keyRef = useRef(0);
  const baseline = useMemo(
    () => ({
      name: initial?.name ?? "",
      description: initial?.description ?? "",
      pairs:
        initial?.sessions
          .map((x) => ({ sessionId: x.session.id, day: x.day }))
          .sort((a, b) => a.day - b.day || a.sessionId - b.sessionId) ?? [],
    }),
    [initial],
  );
  const [name, setName] = useState(baseline.name);
  const [description, setDescription] = useState(baseline.description);
  const [rows, setRows] = useState<Row[]>([]);
  const [step, setStep] = useState(0);
  const [errors, setErrors] = useState<string[]>([]);
  useEffect(() => {
    if (open) {
      setName(baseline.name);
      setDescription(baseline.description);
      setRows(baseline.pairs.map((x) => ({ ...x, key: ++keyRef.current })));
      setStep(0);
      setErrors([]);
    }
  }, [baseline, open]);
  const pairs = rows
    .map(({ sessionId, day }) => ({ sessionId, day }))
    .sort((a, b) => a.day - b.day || a.sessionId - b.sessionId);
  const dirty =
    JSON.stringify({ name, description, pairs }) !== JSON.stringify(baseline);
  useEffect(() => onDirtyChange(open && dirty), [dirty, onDirtyChange, open]);
  async function save() {
    const next: string[] = [];
    if (!name) next.push("Nombre es obligatorio.");
    if (rows.some((x) => x.day < 1)) next.push("Elegí día para cada sesión.");
    const seen = new Set<string>();
    if (
      rows.some((x) => {
        const key = `${x.sessionId}-${x.day}`;
        if (seen.has(key)) return true;
        seen.add(key);
        return false;
      })
    )
      next.push("Una sesión no puede repetirse el mismo día.");
    setErrors(next);
    if (next.length) {
      setStep(next[0].startsWith("Nombre") ? 0 : 1);
      return;
    }
    await onSave({
      name,
      description,
      sessions: rows.map(({ sessionId, day }) => ({ sessionId, day })),
    });
  }
  return (
    <WizardDialog
      open={open}
      title={`${initial ? "Editar" : "Crear"} rutina`}
      activeStep={step}
      onStepChange={setStep}
      dirty={dirty}
      saving={saving}
      saveLabel={initial ? "Guardar cambios" : "Crear rutina"}
      onClose={onClose}
      onSubmit={save}
      steps={[
        {
          label: "Datos básicos",
          hasError: errors.some((x) => x.startsWith("Nombre")),
          content: (
            <div className="grid gap-4">
              <label className="grid gap-2">
                <span>Nombre</span>
                <input
                  className="input input-bordered w-full"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </label>
              <label className="grid gap-2">
                <span>Descripción (opcional)</span>
                <textarea
                  className="textarea textarea-bordered w-full"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </label>
            </div>
          ),
        },
        {
          label: "Sesiones",
          hasError: errors.length > 0,
          content: (
            <div className="grid gap-6">
              {errors.length > 0 && (
                <div className="alert alert-error" role="alert">
                  {errors.join(" ")}
                </div>
              )}
              <SearchableCatalog label="Buscar sesiones" items={sessions}>
                {(session) => (
                  <button
                    type="button"
                    className="btn btn-ghost w-full justify-between"
                    onClick={() =>
                      setRows((current) => [
                        ...current,
                        {
                          key: ++keyRef.current,
                          sessionId: session.id,
                          day: 0,
                        },
                      ])
                    }
                  >
                    <span>{session.name}</span>
                    <span className="inline-flex items-center gap-2">
                      <Plus aria-hidden="true" size={18} />
                      Agregar
                    </span>
                  </button>
                )}
              </SearchableCatalog>
              <div className="grid gap-3">
                {rows.map((row, index) => (
                  <div
                    key={row.key}
                    className="flex flex-wrap items-end gap-3 rounded-box bg-base-200 p-4"
                  >
                    <label className="grid min-w-48 flex-1 gap-2">
                      <span>
                        {sessions.find((x) => x.id === row.sessionId)?.name}
                      </span>
                      <select
                        className="select select-bordered w-full"
                        aria-label={`Día para ${sessions.find((x) => x.id === row.sessionId)?.name} ${index + 1}`}
                        value={row.day}
                        onChange={(e) =>
                          setRows((current) =>
                            current.map((x) =>
                              x.key === row.key
                                ? { ...x, day: Number(e.target.value) }
                                : x,
                            ),
                          )
                        }
                      >
                        <option value="0">Elegir día</option>
                        {days.map((day, i) => (
                          <option value={i + 1} key={day}>
                            {day}
                          </option>
                        ))}
                      </select>
                    </label>
                    <button
                      type="button"
                      className="btn btn-error btn-square"
                      aria-label={`Quitar ${sessions.find((x) => x.id === row.sessionId)?.name}`}
                      onClick={() =>
                        setRows((current) =>
                          current.filter((x) => x.key !== row.key),
                        )
                      }
                    >
                      <X aria-hidden="true" />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          ),
        },
        {
          label: "Resumen",
          content: (
            <div className="card bg-base-200">
              <div className="card-body">
                <h4 className="card-title">{name || "Sin nombre"}</h4>
                <p>{description || "Sin descripción"}</p>
                <ul>
                  {rows.map((row) => (
                    <li key={row.key}>
                      {sessions.find((x) => x.id === row.sessionId)?.name} —{" "}
                      {days[row.day - 1] ?? "Sin día"}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          ),
        },
      ]}
    />
  );
}
