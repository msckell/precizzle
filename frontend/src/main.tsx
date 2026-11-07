import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";

async function start() {
  // Until the expression API is merged, `npm run dev:mock` answers evaluate in the browser (D82).
  if (import.meta.env.MODE === "mock") {
    const { worker } = await import("./mocks/browser.ts");
    await worker.start({ onUnhandledFrame: "bypass" });
  }
  createRoot(document.getElementById("root")!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
}

void start();
