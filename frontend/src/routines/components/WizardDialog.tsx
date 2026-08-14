import {
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { ChevronLeft, ChevronRight, LogOut, Save } from "lucide-react";
import { ModalDialog } from "./ModalDialog";

export interface WizardStep {
  label: string;
  content: ReactNode;
  hasError?: boolean;
}

interface WizardDialogProps {
  open: boolean;
  title: string;
  steps: WizardStep[];
  activeStep: number;
  dirty: boolean;
  saving: boolean;
  saveLabel: string;
  onStepChange: (step: number) => void;
  onClose: () => void;
  onSubmit: () => void;
}

export function WizardDialog(props: WizardDialogProps) {
  const tabsRef = useRef<Array<HTMLButtonElement | null>>([]);
  const [confirmingClose, setConfirmingClose] = useState(false);
  const finalStep = props.steps.length - 1;

  useEffect(() => {
    if (!props.open) setConfirmingClose(false);
  }, [props.open]);

  function requestClose(reason: "button" | "backdrop" | "escape" = "button") {
    if (props.saving) return;
    if (reason === "backdrop" || props.dirty) {
      setConfirmingClose(true);
      return;
    }
    props.onClose();
  }

  function discardChanges() {
    setConfirmingClose(false);
    props.onClose();
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    if (props.activeStep === finalStep && !props.saving) props.onSubmit();
  }

  function moveFocus(index: number) {
    const next = (index + props.steps.length) % props.steps.length;
    props.onStepChange(next);
    tabsRef.current[next]?.focus();
  }

  return (
    <>
      <ModalDialog
        open={props.open}
        title={props.title}
        busy={props.saving}
        onRequestClose={requestClose}
      >
        <form noValidate onSubmit={submit}>
          <div
            role="tablist"
            aria-label={`Pasos de ${props.title}`}
            className="mb-6 flex overflow-x-auto border-b border-base-300"
          >
            {props.steps.map((step, index) => (
              <button
                key={step.label}
                ref={(element) => {
                  tabsRef.current[index] = element;
                }}
                type="button"
                role="tab"
                id={`wizard-tab-${index}`}
                aria-controls={`wizard-panel-${index}`}
                aria-selected={props.activeStep === index}
                tabIndex={props.activeStep === index ? 0 : -1}
                className={`min-h-11 whitespace-nowrap border-b-2 px-4 py-3 ${props.activeStep === index ? "border-secondary text-secondary" : "border-transparent"} ${step.hasError ? "text-error" : ""}`}
                onClick={() => props.onStepChange(index)}
                onKeyDown={(event) => {
                  if (event.key === "ArrowRight") {
                    event.preventDefault();
                    moveFocus(index + 1);
                  }
                  if (event.key === "ArrowLeft") {
                    event.preventDefault();
                    moveFocus(index - 1);
                  }
                  if (event.key === "Home") {
                    event.preventDefault();
                    moveFocus(0);
                  }
                  if (event.key === "End") {
                    event.preventDefault();
                    moveFocus(finalStep);
                  }
                }}
              >
                {index + 1}. {step.label}
                {step.hasError ? " — revisar" : ""}
              </button>
            ))}
          </div>
          {props.steps.map((step, index) => (
            <section
              key={step.label}
              id={`wizard-panel-${index}`}
              role="tabpanel"
              aria-labelledby={`wizard-tab-${index}`}
              hidden={props.activeStep !== index}
            >
              <h3 className="mb-4 text-xl font-semibold">{step.label}</h3>
              {step.content}
            </section>
          ))}
          <div className="mt-8 flex flex-wrap justify-between gap-3 border-t border-base-300 pt-4">
            <button
              type="button"
              className="btn btn-ghost"
              disabled={props.saving}
              onClick={() => requestClose("button")}
            >
              Cancelar
            </button>
            <div className="flex gap-2">
              {props.activeStep > 0 && (
                <button
                  type="button"
                  className="btn btn-secondary"
                  disabled={props.saving}
                  onClick={() => props.onStepChange(props.activeStep - 1)}
                >
                  <ChevronLeft aria-hidden="true" size={18} /> Anterior
                </button>
              )}
              {props.activeStep < finalStep ? (
                <button
                  key="next-step"
                  type="button"
                  className="btn btn-primary"
                  onClick={(event) => {
                    event.preventDefault();
                    props.onStepChange(props.activeStep + 1);
                  }}
                >
                  Siguiente <ChevronRight aria-hidden="true" size={18} />
                </button>
              ) : (
                <button
                  key="save-wizard"
                  type="submit"
                  className="btn btn-primary"
                  disabled={props.saving}
                >
                  <Save aria-hidden="true" size={18} />{" "}
                  {props.saving ? "Guardando…" : props.saveLabel}
                </button>
              )}
            </div>
          </div>
        </form>
      </ModalDialog>
      <ModalDialog
        open={confirmingClose}
        role="alertdialog"
        title="¿Salir del wizard?"
        description="Los cambios se perderán, ¿seguro deseas salir?"
        onRequestClose={() => setConfirmingClose(false)}
      >
        <div className="flex flex-wrap justify-end gap-3">
          <button
            type="button"
            className="btn btn-ghost"
            onClick={() => setConfirmingClose(false)}
          >
            Seguir editando
          </button>
          <button
            type="button"
            className="btn btn-error"
            onClick={discardChanges}
          >
            <LogOut aria-hidden="true" size={18} />
            Salir sin guardar
          </button>
        </div>
      </ModalDialog>
    </>
  );
}
