import { useEffect, useMemo, useState } from "react";
import type { Exercise, ExerciseInput } from "./types";
import { MediaPreview } from "./components/MediaPreview";
import { WizardDialog } from "./components/WizardDialog";

interface ExerciseWizardProps {
  open: boolean;
  initial?: Exercise;
  saving: boolean;
  onClose: () => void;
  onDirtyChange: (dirty: boolean) => void;
  onSave: (input: ExerciseInput) => Promise<void>;
}

const empty: ExerciseInput = {
  name: "",
  description: "",
  imageUrl: "",
  videoUrl: "",
};

export function ExerciseWizard({
  open,
  initial,
  saving,
  onClose,
  onDirtyChange,
  onSave,
}: ExerciseWizardProps) {
  const baseline = useMemo<ExerciseInput>(
    () =>
      initial
        ? {
            name: initial.name,
            description: initial.description ?? "",
            imageUrl: initial.imageUrl ?? "",
            videoUrl: initial.videoUrl ?? "",
          }
        : empty,
    [initial],
  );
  const [draft, setDraft] = useState<ExerciseInput>(baseline);
  const [step, setStep] = useState(0);
  const [errors, setErrors] = useState<string[]>([]);
  const dirty = JSON.stringify(draft) !== JSON.stringify(baseline);

  useEffect(() => {
    if (open) {
      setDraft(baseline);
      setStep(0);
      setErrors([]);
    }
  }, [open, baseline]);
  useEffect(() => onDirtyChange(open && dirty), [dirty, onDirtyChange, open]);

  async function save() {
    const next: string[] = [];
    if (!draft.name) next.push("Nombre es obligatorio.");
    for (const [label, value] of [
      ["URL de imagen", draft.imageUrl],
      ["URL de video", draft.videoUrl],
    ] as const) {
      if (value) {
        try {
          const url = new URL(value);
          if (!["http:", "https:"].includes(url.protocol))
            next.push(`${label} debe usar HTTP o HTTPS.`);
        } catch {
          next.push(`${label} no es válida.`);
        }
      }
    }
    setErrors(next);
    if (next.length) {
      setStep(0);
      return;
    }
    await onSave(draft);
  }

  const field = (key: keyof ExerciseInput, label: string, type = "text") => (
    <label className="grid gap-2">
      <span>{label}</span>
      <input
        className="input input-bordered w-full"
        type={type}
        value={draft[key] ?? ""}
        onChange={(e) => setDraft({ ...draft, [key]: e.target.value })}
      />
    </label>
  );
  return (
    <WizardDialog
      open={open}
      title={`${initial ? "Editar" : "Crear"} ejercicio`}
      activeStep={step}
      onStepChange={setStep}
      dirty={dirty}
      saving={saving}
      saveLabel={initial ? "Guardar cambios" : "Crear ejercicio"}
      onClose={onClose}
      onSubmit={save}
      steps={[
        {
          label: "Datos básicos",
          hasError: errors.length > 0,
          content: (
            <div className="grid gap-4">
              {errors.length > 0 && (
                <div className="alert alert-error" role="alert">
                  {errors.join(" ")}
                </div>
              )}
              {field("name", "Nombre")}
              {
                <label className="grid gap-2">
                  <span>Descripción (opcional)</span>
                  <textarea
                    className="textarea textarea-bordered w-full"
                    value={draft.description}
                    onChange={(e) =>
                      setDraft({ ...draft, description: e.target.value })
                    }
                  />
                </label>
              }
              {field("imageUrl", "URL de imagen (opcional)", "url")}
              {field("videoUrl", "URL de video (opcional)", "url")}
            </div>
          ),
        },
        {
          label: "Resumen",
          content: (
            <div className="grid gap-4">
              <div className="card bg-base-200">
                <div className="card-body">
                  <h4 className="card-title">{draft.name || "Sin nombre"}</h4>
                  <p>{draft.description || "Sin descripción"}</p>
                </div>
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <section
                  className="card min-w-0 bg-base-200"
                  aria-labelledby="exercise-summary-image"
                >
                  <div className="card-body gap-4">
                    <h5
                      id="exercise-summary-image"
                      className="card-title text-lg"
                    >
                      Imagen
                    </h5>
                    <MediaPreview
                      name={draft.name || "ejercicio"}
                      imageUrl={draft.imageUrl}
                      expanded={false}
                    />
                  </div>
                </section>
                <section
                  className="card min-w-0 bg-base-200"
                  aria-labelledby="exercise-summary-video"
                >
                  <div className="card-body gap-4">
                    <h5
                      id="exercise-summary-video"
                      className="card-title text-lg"
                    >
                      Video
                    </h5>
                    <MediaPreview
                      name={draft.name || "ejercicio"}
                      videoUrl={draft.videoUrl}
                      expanded
                      showImage={false}
                    />
                  </div>
                </section>
              </div>
            </div>
          ),
        },
      ]}
    />
  );
}
