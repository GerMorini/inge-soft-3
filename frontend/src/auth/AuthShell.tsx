import { useState, type ReactNode } from "react";
import gymImage from "../assets/fitpro-gym.webp";

interface AuthShellProps {
  title: string;
  children: ReactNode;
}

export function AuthShell({ title, children }: AuthShellProps) {
  const [imageFailed, setImageFailed] = useState(false);
  return (
    <main className="relative min-h-screen overflow-hidden bg-base-200">
      {!imageFailed && (
        <img
          src={gymImage}
          alt=""
          aria-hidden="true"
          className="absolute inset-0 h-full w-full object-cover"
          onError={() => setImageFailed(true)}
        />
      )}
      <div
        className="absolute inset-0 bg-gradient-to-r from-base-200 via-base-200/90 to-base-200/35"
        aria-hidden="true"
      />
      <section
        className="relative flex min-h-screen items-center px-4 py-10 sm:px-8 lg:px-16"
        aria-labelledby="auth-title"
      >
        <div className="card w-full max-w-xl bg-base-100/95 shadow-2xl backdrop-blur-sm">
          <div className="card-body pb-0">
            <p className="text-sm font-bold uppercase tracking-[0.2em] text-primary">
              FitPro
            </p>
            <h1 id="auth-title" className="text-3xl font-bold">
              {title}
            </h1>
          </div>
          {children}
        </div>
      </section>
    </main>
  );
}
