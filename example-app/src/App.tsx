import { useEffect } from "react";
import { useNotesStore } from "./store/useNotesStore";
import { Layout } from "./components/Layout";
import { Sidebar } from "./components/Sidebar";
import { Editor } from "./components/Editor";
import { EmptyState } from "./components/EmptyState";

function App() {
  const initialize = useNotesStore((state) => state.initialize);
  const isLoading = useNotesStore((state) => state.isLoading);
  const isInitialized = useNotesStore((state) => state.isInitialized);
  const error = useNotesStore((state) => state.error);

  useEffect(() => {
    initialize();
  }, [initialize]);

  if (isLoading && !isInitialized) {
    return <EmptyState type="loading" />;
  }

  if (error && !isInitialized) {
    return <EmptyState type="error" message={error} />;
  }

  return (
    <Layout sidebar={<Sidebar />}>
      <Editor />
    </Layout>
  );
}

export default App;
