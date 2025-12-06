import { useState, type ReactNode } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { useIsMobile } from "../../hooks/useMediaQuery";
import styles from "./Layout.module.css";

interface LayoutProps {
  sidebar: ReactNode;
  children: ReactNode;
}

export function Layout({ sidebar, children }: LayoutProps) {
  const isMobile = useIsMobile();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const openSidebar = () => setSidebarOpen(true);
  const closeSidebar = () => setSidebarOpen(false);

  return (
    <div className={styles.layout}>
      {/* Header */}
      <header className={styles.header}>
        {isMobile && (
          <button
            className={styles.menuButton}
            onClick={openSidebar}
            aria-label="Open menu"
          >
            <MenuIcon />
          </button>
        )}
        <h1 className={styles.title}>Notes</h1>
      </header>

      {/* Desktop sidebar */}
      {!isMobile && <aside className={styles.sidebar}>{sidebar}</aside>}

      {/* Mobile sidebar overlay */}
      <AnimatePresence>
        {isMobile && sidebarOpen && (
          <>
            <motion.div
              className={styles.overlay}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              onClick={closeSidebar}
            />
            <motion.aside
              className={styles.mobileSidebar}
              initial={{ x: "-100%" }}
              animate={{ x: 0 }}
              exit={{ x: "-100%" }}
              transition={{ type: "tween", duration: 0.25, ease: "easeOut" }}
            >
              <div className={styles.mobileSidebarHeader}>
                <h2 className={styles.mobileSidebarTitle}>Notes</h2>
                <button
                  className={styles.closeButton}
                  onClick={closeSidebar}
                  aria-label="Close menu"
                >
                  <CloseIcon />
                </button>
              </div>
              <div onClick={closeSidebar}>{sidebar}</div>
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* Main content */}
      <main className={styles.main}>{children}</main>
    </div>
  );
}

function MenuIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
    >
      <path d="M3 5h14M3 10h14M3 15h14" />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
    >
      <path d="M5 5l10 10M15 5L5 15" />
    </svg>
  );
}
