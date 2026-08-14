import { useEffect, useState } from "react";
import { ExternalLink, Image as ImageIcon, Video } from "lucide-react";
import { youtubeEmbedURL } from "./youtube";

interface MediaPreviewProps {
  name: string;
  imageUrl?: string;
  videoUrl?: string;
  expanded?: boolean;
  showImage?: boolean;
  showVideoLink?: boolean;
}

export function MediaPreview({
  name,
  imageUrl,
  videoUrl,
  expanded = true,
  showImage = true,
  showVideoLink = true,
}: MediaPreviewProps) {
  const [imageFailed, setImageFailed] = useState(false);
  const [videoFailed, setVideoFailed] = useState(false);
  const youtubeURL = videoUrl ? youtubeEmbedURL(videoUrl) : undefined;

  useEffect(() => setImageFailed(false), [imageUrl]);
  useEffect(() => setVideoFailed(false), [videoUrl]);

  return (
    <div className="grid gap-4">
      {showImage &&
        (imageUrl && !imageFailed ? (
          <img
            src={imageUrl}
            alt={`Demostración de ${name}`}
            loading="lazy"
            referrerPolicy="no-referrer"
            className="max-h-72 w-full rounded-box object-cover"
            onError={() => setImageFailed(true)}
          />
        ) : (
          <div className="flex min-h-36 items-center justify-center rounded-box bg-base-200 text-base-content/60">
            <ImageIcon aria-hidden="true" size={36} />
            <span className="sr-only">Imagen no disponible</span>
          </div>
        ))}
      {expanded && youtubeURL && (
        <iframe
          src={youtubeURL}
          title={`Video de ejecución de ${name}`}
          loading="lazy"
          referrerPolicy="strict-origin-when-cross-origin"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
          allowFullScreen
          className="aspect-video w-full rounded-box bg-black"
        />
      )}
      {expanded && videoUrl && !youtubeURL && !videoFailed && (
        <video
          src={videoUrl}
          controls
          playsInline
          preload="metadata"
          className="max-h-[32rem] w-full rounded-box bg-black"
          onError={() => setVideoFailed(true)}
        />
      )}
      {expanded && videoUrl && !youtubeURL && videoFailed && (
        <p className="alert alert-warning">
          <Video aria-hidden="true" size={20} />
          No se pudo reproducir el video.
        </p>
      )}
      {expanded && !showImage && !videoUrl && (
        <p className="flex min-h-36 items-center justify-center gap-2 rounded-box bg-base-200 text-base-content/70">
          <Video aria-hidden="true" size={24} />
          Video no disponible.
        </p>
      )}
      {showVideoLink && videoUrl && (
        <a
          className="link link-secondary inline-flex items-center gap-2"
          href={videoUrl}
          target="_blank"
          rel="noopener noreferrer"
        >
          <ExternalLink aria-hidden="true" size={18} />
          Abrir video original de {name}
        </a>
      )}
    </div>
  );
}
