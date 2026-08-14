import { fireEvent, render, screen } from "@testing-library/react";
import { MediaPreview } from "./MediaPreview";

it("renders native deferred video and permanent safe link", () => {
  const { container } = render(
    <MediaPreview name="Remo" videoUrl="https://example.test/video" />,
  );
  const video = container.querySelector("video");
  expect(video).toHaveAttribute("controls");
  expect(video).toHaveAttribute("preload", "metadata");
  expect(container.querySelector("iframe")).toBeNull();
  const link = screen.getByRole("link", { name: /Abrir video original/ });
  expect(link).toHaveAttribute("rel", "noopener noreferrer");
  fireEvent.error(video!);
  expect(
    screen.getByText("No se pudo reproducir el video."),
  ).toBeInTheDocument();
  expect(link).toBeInTheDocument();
});

it.each([
  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "https://youtu.be/dQw4w9WgXcQ?t=30",
  "https://m.youtube.com/shorts/dQw4w9WgXcQ",
  "https://www.youtube.com/live/dQw4w9WgXcQ",
])("embeds supported YouTube URL %s", (videoUrl) => {
  const { container } = render(
    <MediaPreview name="Remo" videoUrl={videoUrl} />,
  );

  const player = screen.getByTitle("Video de ejecución de Remo");
  expect(player).toHaveAttribute(
    "src",
    "https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ",
  );
  expect(player).toHaveAttribute(
    "referrerpolicy",
    "strict-origin-when-cross-origin",
  );
  expect(player).toHaveAttribute("allowfullscreen");
  expect(container.querySelector("video")).toBeNull();
  expect(
    screen.getByRole("link", { name: /Abrir video original/ }),
  ).toHaveAttribute("href", videoUrl);
});

it("does not embed a lookalike YouTube domain", () => {
  const { container } = render(
    <MediaPreview
      name="Remo"
      videoUrl="https://youtube.com.example.test/watch?v=dQw4w9WgXcQ"
    />,
  );

  expect(container.querySelector("iframe")).toBeNull();
  expect(container.querySelector("video")).toBeInTheDocument();
});

it("shows a video fallback without restoring the image", () => {
  render(<MediaPreview name="Remo" expanded showImage={false} />);

  expect(screen.getByText("Video no disponible.")).toBeInTheDocument();
  expect(screen.queryByText("Imagen no disponible")).not.toBeInTheDocument();
});
