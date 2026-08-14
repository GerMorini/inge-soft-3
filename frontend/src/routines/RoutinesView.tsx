import { useCallback, useEffect, useState } from "react";
import { CalendarDays, Eye, Pencil, Plus, Trash2 } from "lucide-react";
import { ApiError } from "../auth/types";
import {
  createRoutine,
  deleteRoutine,
  getRoutine,
  listRoutines,
  listSessions,
  updateRoutine,
} from "./api";
import { RoutineWizard } from "./RoutineWizard";
import { FeatureHero } from "./components/FeatureHero";
import { MediaPreview } from "./components/MediaPreview";
import { ModalDialog } from "./components/ModalDialog";
import type {
  RoutineDetail,
  RoutineInput,
  RoutineSummary,
  SessionExercise,
  SessionSummary,
} from "./types";

const days = [
  "Lunes",
  "Martes",
  "Miércoles",
  "Jueves",
  "Viernes",
  "Sábado",
  "Domingo",
];
interface Props {
  onUnauthenticated: () => void;
  onDirtyChange?: (dirty: boolean) => void;
}

export function RoutinesView({
  onUnauthenticated,
  onDirtyChange = () => undefined,
}: Props) {
  const [items, setItems] = useState<RoutineSummary[]>([]);
  const [sessions, setSessions] = useState<SessionSummary[]>([]);
  const [detail, setDetail] = useState<RoutineDetail>();
  const [editing, setEditing] = useState<RoutineDetail>();
  const [wizardOpen, setWizardOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const load = useCallback(() => {
    void Promise.all([
      listRoutines(onUnauthenticated),
      listSessions(onUnauthenticated),
    ])
      .then(([r, s]) => {
        setItems(r);
        setSessions(s);
      })
      .catch((e) =>
        setMessage(
          e instanceof ApiError
            ? e.message
            : "No se pudieron cargar las rutinas.",
        ),
      );
  }, [onUnauthenticated]);
  useEffect(load, [load]);
  async function fetchDetail(item: RoutineSummary) {
    const found = await getRoutine(item.id, onUnauthenticated);
    setDetail(found);
    return found;
  }
  async function view(item: RoutineSummary) {
    try {
      await fetchDetail(item);
      setDetailOpen(true);
    } catch {
      setMessage("La rutina ya no está disponible.");
    }
  }
  async function edit(item: RoutineSummary) {
    try {
      setEditing(await fetchDetail(item));
      setDetailOpen(false);
      setWizardOpen(true);
    } catch {
      setMessage("La rutina ya no está disponible.");
    }
  }
  function closeWizard() {
    setWizardOpen(false);
    setEditing(undefined);
    onDirtyChange(false);
  }
  async function save(input: RoutineInput) {
    setSaving(true);
    try {
      const saved = editing
        ? await updateRoutine(editing.id, input, onUnauthenticated)
        : await createRoutine(input, onUnauthenticated);
      const summary = {
        id: saved.id,
        name: saved.name,
        description: saved.description,
      };
      setItems((current) =>
        editing
          ? current.map((x) => (x.id === saved.id ? summary : x))
          : [...current, summary],
      );
      setMessage(`Rutina ${editing ? "actualizada" : "creada"}.`);
      closeWizard();
    } catch (e) {
      setMessage(
        e instanceof ApiError ? e.message : "No se pudo guardar la rutina.",
      );
    } finally {
      setSaving(false);
    }
  }
  async function remove(item: RoutineSummary) {
    if (!window.confirm(`¿Eliminar ${item.name}?`)) return;
    try {
      await deleteRoutine(item.id, onUnauthenticated);
      setItems((current) => current.filter((x) => x.id !== item.id));
      setDetailOpen(false);
    } catch (e) {
      setMessage(
        e instanceof ApiError ? e.message : "No se pudo eliminar la rutina.",
      );
    }
  }
  return (
    <section>
      <FeatureHero
        title="Rutinas"
        description="Visualiza y ajusta tus rutinas"
        icon={CalendarDays}
      />
      <div className="mb-6 flex flex-wrap justify-between gap-4">
        <h2 className="text-2xl font-bold">Tus rutinas</h2>
        <button
          className="btn btn-primary"
          type="button"
          onClick={() => {
            setEditing(undefined);
            setWizardOpen(true);
          }}
        >
          <Plus aria-hidden="true" size={18} />
          Crear rutina
        </button>
      </div>
      {message && (
        <div className="alert mb-5" role="status">
          {message}
        </div>
      )}
      <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        {items.map((item) => (
          <article key={item.id} className="card bg-base-100 shadow-lg">
            <div className="card-body">
              <h3 className="card-title">{item.name}</h3>
              <p>{item.description || "Sin descripción."}</p>
              <div className="card-actions mt-4 justify-end">
                <button
                  className="btn btn-secondary"
                  type="button"
                  aria-label={`Ver ${item.name}`}
                  onClick={() => void view(item)}
                >
                  <Eye aria-hidden="true" size={18} />
                  Ver
                </button>
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
          </article>
        ))}
      </div>
      <RoutineWizard
        open={wizardOpen}
        initial={editing}
        sessions={sessions}
        saving={saving}
        onClose={closeWizard}
        onDirtyChange={onDirtyChange}
        onSave={save}
      />
      <ModalDialog
        open={detailOpen}
        title={detail?.name ?? "Detalle de rutina"}
        onRequestClose={() => setDetailOpen(false)}
      >
        {detail && (
          <div className="grid gap-4">
            <p>{detail.description || "Sin descripción."}</p>
            {detail.sessions.map((assignment) => (
              <details
                key={`${assignment.session.id}-${assignment.day}`}
                className="collapse-arrow collapse bg-base-200"
              >
                <summary className="collapse-title">
                  <strong>
                    {days[assignment.day - 1]} · {assignment.session.name}
                  </strong>
                  <p>{assignment.session.description || "Sin descripción."}</p>
                </summary>
                <div className="collapse-content grid gap-3">
                  {assignment.session.exercises.map((item) => (
                    <ExerciseDisclosure key={item.exercise.id} item={item} />
                  ))}
                </div>
              </details>
            ))}
          </div>
        )}
      </ModalDialog>
    </section>
  );
}

function ExerciseDisclosure({ item }: { item: SessionExercise }) {
  const [expanded, setExpanded] = useState(false);
  return (
    <details
      className="collapse-arrow collapse bg-base-100"
      onToggle={(event) => setExpanded(event.currentTarget.open)}
    >
      <summary className="collapse-title flex items-center gap-3">
        {item.exercise.imageUrl && (
          <img
            src={item.exercise.imageUrl}
            alt=""
            loading="lazy"
            referrerPolicy="no-referrer"
            className="h-12 w-12 rounded object-cover"
          />
        )}
        <span>
          {item.exercise.name} · {item.series} × {item.repetitions}
        </span>
      </summary>
      <div className="collapse-content">
        <p className="mb-4">
          {item.exercise.description || "Sin descripción."}
        </p>
        <MediaPreview
          name={item.exercise.name}
          imageUrl={item.exercise.imageUrl}
          videoUrl={item.exercise.videoUrl}
          expanded={expanded}
        />
      </div>
    </details>
  );
}
