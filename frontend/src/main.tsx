import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { ThemeProvider } from "@/modules/theme";
import { usePreferencesStore } from "@/modules/settings/store";
import "./index.css";

// Hydrate persisted preferences (theme, zoom, …) from the backend before the
// first render, then let the ThemeProvider adopt them.
void usePreferencesStore.getState().init();

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ThemeProvider>
      <App />
    </ThemeProvider>
  </React.StrictMode>
);
