/**
 * Screenshot slot. Pass `src` (a file under /public/shots) to render the real
 * capture; without it a dashed placeholder with the label is shown.
 */
export default function Shot({
  src,
  label,
  alt,
  className = "",
}: {
  src?: string;
  label: string;
  alt?: string;
  className?: string;
}) {
  if (src) {
    return (
      <img
        src={src}
        alt={alt ?? label}
        className={`block h-full w-full object-cover object-top ${className}`}
        loading="lazy"
      />
    );
  }
  return (
    <div className={`shot ${className}`} aria-label={label}>
      <span className="whitespace-pre-line px-3">{label}</span>
    </div>
  );
}
