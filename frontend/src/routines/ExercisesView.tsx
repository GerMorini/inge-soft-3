import { useCallback, useEffect, useState } from "react";
import {
  ChevronDown,
  Dumbbell,
  ExternalLink,
  Pencil,
  Plus,
  Trash2,
} from "lucide-react";
import { ApiError } from "../auth/types";
import {
  createExercise,
  deleteExercise,
  listExercises,
  updateExercise,
} from "./api";
import { ExerciseWizard } from "./ExerciseWizard";
import { FeatureHero } from "./components/FeatureHero";
import { MediaPreview } from "./components/MediaPreview";
import type { Exercise, ExerciseInput } from "./types";

interface Props {
  onUnauthenticated: () => void;
  onDirtyChange?: (dirty: boolean) => void;
}

export function ExercisesView({
  onUnauthenticated,
  onDirtyChange = () => undefined,
}: Props) {
  const [items, setItems] = useState<Exercise[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [mode, setMode] = useState<"closed" | "create" | "edit">("closed");
  const [editing, setEditing] = useState<Exercise>();
  const [saving, setSaving] = useState(false);
  const [expanded, setExpanded] = useState<Set<number>>(() => new Set());
  const load = useCallback(() => {
    setLoading(true);
    void listExercises(onUnauthenticated)
      .then(setItems)
      .catch((e) =>
        setMessage(
          e instanceof ApiError
            ? e.message
            : "No se pudieron cargar los ejercicios.",
        ),
      )
      .finally(() => setLoading(false));
  }, [onUnauthenticated]);
  useEffect(load, [load]);
  function close() {
    setMode("closed");
    setEditing(undefined);
    onDirtyChange(false);
  }
  async function save(input: ExerciseInput) {
    setSaving(true);
    setMessage("");
    try {
      const saved = editing
        ? await updateExercise(editing.id, input, onUnauthenticated)
        : await createExercise(input, onUnauthenticated);
      setItems((current) =>
        editing
          ? current.map((x) => (x.id === saved.id ? saved : x))
          : [...current, saved],
      );
      setMessage(`Ejercicio ${editing ? "actualizado" : "creado"}.`);
      close();
    } catch (e) {
      setMessage(
        e instanceof ApiError ? e.message : "No se pudo guardar el ejercicio.",
      );
    } finally {
      setSaving(false);
    }
  }
  async function remove(item: Exercise) {
    if (!window.confirm(`¿Eliminar ${item.name}?`)) return;
    try {
      await deleteExercise(item.id, onUnauthenticated);
      setItems((current) => current.filter((x) => x.id !== item.id));
    } catch (e) {
      setMessage(
        e instanceof ApiError ? e.message : "No se pudo eliminar el ejercicio.",
      );
    }
  }
  return (
    <section>
      <FeatureHero
        title="Ejercicios"
        description="Crea y consulta ejercicios para tus sesiones"
        icon={Dumbbell}
      />
      <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <h2 className="text-2xl font-bold">Tus ejercicios</h2>
        <button
          className="btn btn-primary"
          type="button"
          onClick={() => {
            setEditing(undefined);
            setMode("create");
          }}
        >
          <Plus aria-hidden="true" size={18} />
          Crear ejercicio
        </button>
      </div>
      {message && (
        <div className="alert mb-5" role="status">
          {message}
        </div>
      )}
      {loading ? (
        <p role="status">Cargando ejercicios…</p>
      ) : items.length === 0 ? (
        <p className="rounded-box bg-base-100 p-8 text-center">
          Todavía no creaste ejercicios.
        </p>
      ) : (
        <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          {items.map((item) => {
            const isExpanded = expanded.has(item.id);
            const contentID = `exercise-detail-${item.id}`;
            const toggle = () =>
              setExpanded((current) => {
                const next = new Set(current);
                if (next.has(item.id)) next.delete(item.id);
                else next.add(item.id);
                return next;
              });
            const cardTitle = (
              <span className="flex items-center justify-between gap-3 px-5 py-4">
                <h3 className="text-xl font-semibold">{item.name}</h3>
                <ChevronDown
                  aria-hidden="true"
                  size={20}
                  className={`shrink-0 transition-transform ${isExpanded ? "rotate-180" : ""}`}
                />
              </span>
            );
            return (
              <article
                key={item.id}
                className="card overflow-hidden bg-base-100 shadow-lg"
              >
                {isExpanded ? (
                  <>
                    <MediaPreview
                      name={item.name}
                      videoUrl={item.videoUrl}
                      expanded
                      showImage={false}
                      showVideoLink={false}
                    />
                    <button
                      type="button"
                      className="w-full text-left"
                      aria-label={`Ocultar detalles de ${item.name}`}
                      aria-expanded="true"
                      aria-controls={contentID}
                      onClick={toggle}
                    >
                      {cardTitle}
                    </button>
                  </>
                ) : (
                  <button
                    type="button"
                    className="w-full text-left"
                    aria-label={`Ver detalles de ${item.name}`}
                    aria-expanded="false"
                    aria-controls={contentID}
                    onClick={toggle}
                  >
                    <MediaPreview
                      name={item.name}
                      imageUrl={item.imageUrl}
                      expanded={false}
                    />
                    {cardTitle}
                  </button>
                )}
                {isExpanded && (
                  <div id={contentID} className="grid gap-3 px-5 pb-5">
                    <p>{item.description || "Sin descripción."}</p>
                    {item.videoUrl && (
                      <a
                        className="link link-secondary inline-flex items-center gap-2"
                        href={item.videoUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <ExternalLink aria-hidden="true" size={18} />
                        Abrir video original de {item.name}
                      </a>
                    )}
                  </div>
                )}
                <div className="card-actions mt-auto justify-end border-t border-base-300 p-4">
                  <button
                    type="button"
                    className="btn btn-secondary"
                    aria-label={`Editar ${item.name}`}
                    onClick={() => {
                      setEditing(item);
                      setMode("edit");
                    }}
                  >
                    <Pencil aria-hidden="true" size={18} />
                    Editar
                  </button>
                  <button
                    type="button"
                    className="btn btn-error"
                    aria-label={`Eliminar ${item.name}`}
                    onClick={() => void remove(item)}
                  >
                    <Trash2 aria-hidden="true" size={18} />
                    Eliminar
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      )}
      <ExerciseWizard
        open={mode !== "closed"}
        initial={editing}
        saving={saving}
        onDirtyChange={onDirtyChange}
        onClose={close}
        onSave={save}
      />
    </section>
  );
}
