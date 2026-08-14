import { useCallback, useEffect, useState } from "react";
import { ListChecks, Pencil, Plus, Trash2 } from "lucide-react";
import { ApiError } from "../auth/types";
import {
  createSession,
  deleteSession,
  getSession,
  listExercises,
  listSessions,
  updateSession,
} from "./api";
import { SessionWizard } from "./SessionWizard";
import { FeatureHero } from "./components/FeatureHero";
import type {
  Exercise,
  SessionDetail,
  SessionInput,
  SessionSummary,
} from "./types";

interface Props {
  onUnauthenticated: () => void;
  onDirtyChange?: (dirty: boolean) => void;
}
export function SessionsView({
  onUnauthenticated,
  onDirtyChange = () => undefined,
}: Props) {
  const [items, setItems] = useState<SessionSummary[]>([]);
  const [catalog, setCatalog] = useState<Exercise[]>([]);
  const [details, setDetails] = useState<Record<number, SessionDetail>>({});
  const [editing, setEditing] = useState<SessionDetail>();
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const load = useCallback(() => {
    void Promise.all([
      listSessions(onUnauthenticated),
      listExercises(onUnauthenticated),
    ])
      .then(([sessions, exercises]) => {
        setItems(sessions);
        setCatalog(exercises);
      })
      .catch((e) =>
        setMessage(
          e instanceof ApiError
            ? e.message
            : "No se pudieron cargar las sesiones.",
        ),
      );
  }, [onUnauthenticated]);
  useEffect(load, [load]);
  async function detail(id: number) {
    if (details[id]) return details[id];
    const found = await getSession(id, onUnauthenticated);
    setDetails((current) => ({ ...current, [id]: found }));
    return found;
  }
  async function edit(item: SessionSummary) {
    try {
      setEditing(await detail(item.id));
      setOpen(true);
    } catch {
      setMessage("La sesión ya no está disponible.");
    }
  }
  function close() {
    setOpen(false);
    setEditing(undefined);
    onDirtyChange(false);
  }
  async function save(input: SessionInput) {
    setSaving(true);
    try {
      const saved = editing
        ? await updateSession(editing.id, input, onUnauthenticated)
        : await createSession(input, onUnauthenticated);
      const summary = {
        id: saved.id,
        name: saved.name,
        description: saved.description,
        exerciseCount: saved.exercises.length,
      };
      setItems((current) =>
        editing
          ? current.map((x) => (x.id === saved.id ? summary : x))
          : [...current, summary],
      );
      setDetails((current) => ({ ...current, [saved.id]: saved }));
      setMessage(`Sesión ${editing ? "actualizada" : "creada"}.`);
      close();
    } catch (e) {
      setMessage(
        e instanceof ApiError ? e.message : "No se pudo guardar la sesión.",
      );
    } finally {
      setSaving(false);
    }
  }
  async function remove(item: SessionSummary) {
    if (!window.confirm(`¿Eliminar ${item.name}?`)) return;
    try {
      await deleteSession(item.id, onUnauthenticated);
      setItems((current) => current.filter((x) => x.id !== item.id));
    } catch (e) {
      setMessage(
        e instanceof ApiError ? e.message : "No se pudo eliminar la sesión.",
      );
    }
  }
  return (
    <section>
      <FeatureHero
        title="Sesiones"
        description="Organiza ejercicios y define su orden de ejecución"
        icon={ListChecks}
      />
      <div className="mb-6 flex flex-wrap justify-between gap-4">
        <h2 className="text-2xl font-bold">Tus sesiones</h2>
        <button
          className="btn btn-primary"
          type="button"
          onClick={() => {
            setEditing(undefined);
            setOpen(true);
          }}
        >
          <Plus aria-hidden="true" size={18} />
          Crear sesión
        </button>
      </div>
      {message && (
        <div className="alert mb-5" role="status">
          {message}
        </div>
      )}
      <div className="grid gap-4">
        {items.map((item) => (
          <details
            key={item.id}
            className="collapse-arrow collapse bg-base-100"
            onToggle={(e) => {
              if (e.currentTarget.open)
                void detail(item.id).catch(() =>
                  setMessage("No se pudo cargar la sesión."),
                );
            }}
          >
            <summary className="collapse-title">
              <h3 className="text-xl font-semibold">{item.name}</h3>
              <p>
                {item.description || "Sin descripción"} · {item.exerciseCount}{" "}
                ejercicios
              </p>
            </summary>
            <div className="collapse-content">
              <ol className="grid gap-2">
                {details[item.id]?.exercises.map((x) => (
                  <li
                    key={x.exercise.id}
                    className="rounded-box bg-base-200 p-3"
                  >
                    {x.order}. {x.exercise.name} — {x.series} series ×{" "}
                    {x.repetitions} repeticiones
                  </li>
                ))}
              </ol>
              <div className="mt-4 flex justify-end gap-2">
                <button
                  className="btn btn-secondary"
                  type="button"
                  aria-label={`Editar ${item.name}`}
                  onClick={() => void edit(item)}
                >
                  <Pencil aria-hidden="true" size={18} />
                  Editar
                </button>
                <button
                  className="btn btn-error"
                  type="button"
                  aria-label={`Eliminar ${item.name}`}
                  onClick={() => void remove(item)}
                >
                  <Trash2 aria-hidden="true" size={18} />
                  Eliminar
                </button>
              </div>
            </div>
          </details>
        ))}
      </div>
      <SessionWizard
        open={open}
        initial={editing}
        exercises={catalog}
        saving={saving}
        onClose={close}
        onDirtyChange={onDirtyChange}
        onSave={save}
      />
    </section>
  );
}
