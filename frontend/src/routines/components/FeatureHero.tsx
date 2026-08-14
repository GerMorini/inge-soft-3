import type { LucideIcon } from "lucide-react";
import gymImage from "../../assets/fitpro-gym.webp";

export function FeatureHero({
  title,
  description,
  icon: Icon,
}: {
  title: string;
  description: string;
  icon: LucideIcon;
}) {
  return (
    <header className="relative mb-8 min-h-56 overflow-hidden rounded-box bg-base-100">
      <img
        src={gymImage}
        alt=""
        aria-hidden="true"
        className="absolute inset-0 h-full w-full object-cover"
      />
      <div
        className="absolute inset-0 bg-gradient-to-r from-base-200 via-base-200/85 to-transparent"
        aria-hidden="true"
      />
      <div className="relative flex min-h-56 max-w-xl flex-col justify-center p-8">
        <Icon aria-hidden="true" size={36} className="mb-4 text-primary" />
        <h1 className="text-4xl font-bold">{title}</h1>
        <p className="mt-3 text-lg text-base-content/80">{description}</p>
      </div>
    </header>
  );
}
