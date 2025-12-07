import styles from "./EmptyState.module.css";

interface EmptyStateProps {
  type: "no-selection" | "loading" | "error";
  message?: string;
}

export function EmptyState({ type, message }: EmptyStateProps) {
  if (type === "loading") {
    return (
      <div className={styles.container}>
        <div className={styles.spinner} />
        <p className={styles.text}>Loading...</p>
      </div>
    );
  }

  if (type === "error") {
    return (
      <div className={styles.container}>
        <div className={styles.icon}>
          <ErrorIcon />
        </div>
        <p className={styles.text}>{message || "Something went wrong"}</p>
      </div>
    );
  }

  // no-selection
  return (
    <div className={styles.container}>
      <div className={styles.icon}>
        <NoteIcon />
      </div>
      <p className={styles.text}>Select a note</p>
      <p className={styles.hint}>Choose a note from the list or create a new one</p>
    </div>
  );
}

function NoteIcon() {
  return (
    <svg
      width="48"
      height="48"
      viewBox="0 0 48 48"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="8" y="6" width="32" height="36" rx="3" />
      <path d="M14 14h20M14 22h20M14 30h12" />
    </svg>
  );
}

function ErrorIcon() {
  return (
    <svg
      width="48"
      height="48"
      viewBox="0 0 48 48"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="24" cy="24" r="18" />
      <path d="M24 16v10M24 32h.01" />
    </svg>
  );
}
