const videoIDPattern = /^[A-Za-z0-9_-]{11}$/;

export function youtubeEmbedURL(value: string): string | undefined {
  let url: URL;
  try {
    url = new URL(value);
  } catch {
    return undefined;
  }

  if (url.protocol !== "http:" && url.protocol !== "https:") return undefined;

  const hostname = url.hostname.toLowerCase().replace(/^www\./, "");
  let videoID: string | undefined;

  if (hostname === "youtu.be") {
    videoID = url.pathname.split("/").filter(Boolean)[0];
  } else if (
    hostname === "youtube.com" ||
    hostname === "m.youtube.com" ||
    hostname === "music.youtube.com"
  ) {
    if (url.pathname === "/watch")
      videoID = url.searchParams.get("v") ?? undefined;
    else {
      const [kind, id] = url.pathname.split("/").filter(Boolean);
      if (["embed", "shorts", "live", "v"].includes(kind)) videoID = id;
    }
  } else if (hostname === "youtube-nocookie.com") {
    const [kind, id] = url.pathname.split("/").filter(Boolean);
    if (kind === "embed") videoID = id;
  }

  if (!videoIDPattern.test(videoID ?? "")) return undefined;
  return `https://www.youtube-nocookie.com/embed/${videoID}`;
}
