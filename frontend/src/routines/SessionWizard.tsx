import { useEffect, useMemo, useState } from "react";
import { ArrowDown, ArrowUp, X } from "lucide-react";
import type { Exercise, SessionDetail, SessionInput } from "./types";
import { SearchableCatalog } from "./components/SearchableCatalog";
import { WizardDialog } from "./components/WizardDialog";

interface Props {
  open: boolean;
  initial?: SessionDetail;
  exercises: Exercise[];
  saving: boolean;
  onClose: () => void;
  onDirtyChange: (dirty: boolean) => void;
  onSave: (input: SessionInput) => Promise<void>;
}
type Row = { exerciseId: number; series: number; repetitions: number };

export function SessionWizard({
  open,
  initial,
  exercises,
  saving,
  onClose,
  onDirtyChange,
  onSave,
}: Props) {
  const baseline = useMemo(
    () => ({
      name: initial?.name ?? "",
      description: initial?.description ?? "",
      rows:
        initial?.exercises.map((x) => ({
          exerciseId: x.exercise.id,
          series: x.series,
          repetitions: x.repetitions,
        })) ?? [],
    }),
    [initial],
  );
  const [name, setName] = useState(baseline.name);
  const [description, setDescription] = useState(baseline.description);
  const [rows, setRows] = useState<Row[]>(baseline.rows);
  const [step, setStep] = useState(0);
  const [errors, setErrors] = useState<string[]>([]);
  const dirty =
    JSON.stringify({ name, description, rows }) !== JSON.stringify(baseline);
  useEffect(() => {
    if (open) {
      setName(baseline.name);
      setDescription(baseline.description);
      setRows(baseline.rows);
      setStep(0);
      setErrors([]);
    }
  }, [baseline, open]);
  useEffect(() => onDirtyChange(open && dirty), [dirty, onDirtyChange, open]);
  const move = (index: number, offset: number) =>
    setRows((current) => {
      const next = [...current];
      [next[index], next[index + offset]] = [next[index + offset], next[index]];
      return next;
    });
  async function save() {
    const next = name ? [] : ["Nombre es obligatorio."];
    setErrors(next);
    if (next.length) {
      setStep(0);
      return;
    }
    await onSave({
      name,
      description,
      exercises: rows.map((row, index) => ({ ...row, order: index + 1 })),
    });
  }
  return (
    <WizardDialog
      open={open}
      title={`${initial ? "Editar" : "Crear"} sesión`}
      activeStep={step}
      onStepChange={setStep}
      dirty={dirty}
      saving={saving}
      saveLabel={initial ? "Guardar cambios" : "Crear sesión"}
      onClose={onClose}
      onSubmit={save}
      steps={[
        {
          label: "Datos básicos",
          hasError: errors.length > 0,
          content: (
            <div className="grid gap-4">
              {errors.length > 0 && (
                <div role="alert" className="alert alert-error">
                  {errors.join(" ")}
                </div>
              )}
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
          label: "Ejercicios",
          content: (
            <div className="grid gap-6">
              <SearchableCatalog label="Buscar ejercicios" items={exercises}>
                {(exercise) => (
                  <label className="flex min-h-11 items-center gap-3 rounded-box bg-base-200 p-3">
                    <input
                      type="checkbox"
                      className="checkbox checkbox-secondary"
                      checked={rows.some((r) => r.exerciseId === exercise.id)}
                      onChange={(e) =>
                        setRows((current) =>
                          e.target.checked
                            ? [
                                ...current,
                                {
                                  exerciseId: exercise.id,
                                  series: 0,
                                  repetitions: 0,
                                },
                              ]
                            : current.filter(
                                (r) => r.exerciseId !== exercise.id,
                              ),
                        )
                      }
                    />
                    {exercise.name}
                  </label>
                )}
              </SearchableCatalog>
              <div className="grid gap-3">
                {rows.map((row, index) => {
                  const exercise = exercises.find(
                    (x) => x.id === row.exerciseId,
                  );
                  return (
                    <div key={row.exerciseId} className="card bg-base-200">
                      <div className="card-body p-4">
                        <strong>
                          {index + 1}. {exercise?.name}
                        </strong>
                        <div className="grid gap-3 sm:grid-cols-2">
                          <label className="grid gap-2">
                            <span>Series</span>
                            <input
                              className="input input-bordered w-full"
                              type="number"
                              min="0"
                              step="1"
                              value={row.series}
                              onChange={(e) =>
                                setRows((current) =>
                                  current.map((x, i) =>
                                    i === index
                                      ? { ...x, series: Number(e.target.value) }
                                      : x,
                                  ),
                                )
                              }
                            />
                          </label>
                          <label className="grid gap-2">
                            <span>Repeticiones</span>
                            <input
                              className="input input-bordered w-full"
                              type="number"
                              min="0"
                              step="1"
                              value={row.repetitions}
                              onChange={(e) =>
                                setRows((current) =>
                                  current.map((x, i) =>
                                    i === index
                                      ? {
                                          ...x,
                                          repetitions: Number(e.target.value),
                                        }
                                      : x,
                                  ),
                                )
                              }
                            />
                          </label>
                        </div>
                        <div className="card-actions">
                          <button
                            type="button"
                            className="btn btn-sm"
                            aria-label={`Mover ${exercise?.name} arriba`}
                            disabled={index === 0}
                            onClick={() => move(index, -1)}
                          >
                            <ArrowUp aria-hidden="true" />
                          </button>
                          <button
                            type="button"
                            className="btn btn-sm"
                            aria-label={`Mover ${exercise?.name} abajo`}
                            disabled={index === rows.length - 1}
                            onClick={() => move(index, 1)}
                          >
                            <ArrowDown aria-hidden="true" />
                          </button>
                          <button
                            type="button"
                            className="btn btn-sm btn-error"
                            aria-label={`Quitar ${exercise?.name}`}
                            onClick={() =>
                              setRows((current) =>
                                current.filter((_, i) => i !== index),
                              )
                            }
                          >
                            <X aria-hidden="true" />
                          </button>
                        </div>
                      </div>
                    </div>
                  );
                })}
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
                <p>{rows.length} ejercicios</p>
                <ol className="list-decimal pl-5">
                  {rows.map((row) => (
                    <li key={row.exerciseId}>
                      {exercises.find((x) => x.id === row.exerciseId)?.name}:{" "}
                      {row.series} × {row.repetitions}
                    </li>
                  ))}
                </ol>
              </div>
            </div>
          ),
        },
      ]}
    />
  );
}
