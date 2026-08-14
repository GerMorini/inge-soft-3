import { useEffect, useId, useRef, type ReactNode } from "react";
import { X } from "lucide-react";

interface ModalDialogProps {
  open: boolean;
  title: string;
  description?: string;
  children: ReactNode;
  busy?: boolean;
  role?: "dialog" | "alertdialog";
  onRequestClose: (reason: "button" | "backdrop" | "escape") => void;
}

export function ModalDialog({
  open,
  title,
  description,
  children,
  busy = false,
  role = "dialog",
  onRequestClose,
}: ModalDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const titleRef = useRef<HTMLHeadingElement>(null);
  const triggerRef = useRef<HTMLElement | null>(null);
  const titleId = useId();
  const descriptionId = useId();

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) {
      triggerRef.current =
        document.activeElement instanceof HTMLElement
          ? document.activeElement
          : null;
      dialog.showModal();
      requestAnimationFrame(() => titleRef.current?.focus());
    }
    if (!open && dialog.open) {
      dialog.close();
      requestAnimationFrame(() => triggerRef.current?.focus());
    }
  }, [open]);

  return (
    <dialog
      ref={dialogRef}
      role={role}
      className="modal"
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
      aria-busy={busy}
      onCancel={(event) => {
        event.preventDefault();
        if (!busy) onRequestClose("escape");
      }}
    >
      <div className="modal-box max-h-[90vh] max-w-4xl overflow-y-auto bg-base-100">
        <button
          type="button"
          className="btn btn-ghost btn-square absolute right-3 top-3 min-h-11 min-w-11"
          aria-label={`Cerrar ${title}`}
          disabled={busy}
          onClick={() => onRequestClose("button")}
        >
          <X aria-hidden="true" size={20} />
        </button>
        <h2
          ref={titleRef}
          tabIndex={-1}
          id={titleId}
          className="pr-12 text-2xl font-bold"
        >
          {title}
        </h2>
        {description && (
          <p id={descriptionId} className="mt-2 text-base-content/70">
            {description}
          </p>
        )}
        <div className="mt-6">{children}</div>
      </div>
      <button
        type="button"
        className="modal-backdrop cursor-default"
        aria-label={`Cerrar ${title}`}
        disabled={busy}
        onClick={() => onRequestClose("backdrop")}
      />
    </dialog>
  );
}
