import { isAIEnabled, streamAIInsights } from "../../api";
import { requiredEl } from "../ui";

export function renderAI(): void {
  const aiPanel = requiredEl<HTMLElement>("ai-panel");
  isAIEnabled().then((enabled) => {
    if (!enabled) return;
    aiPanel.classList.remove("hidden");
    aiPanel.innerHTML = `<h3>✨ Virtual consultant insights</h3><div id="ai-stream" class="skeleton">Loading insights…</div>`;
    let text = "";
    streamAIInsights(
      (chunk) => {
        text += chunk;
        const el = requiredEl<HTMLElement>("ai-stream");
        el.classList.remove("skeleton");
        el.textContent = text;
      },
      () => {},
      (msg) => {
        const el = document.getElementById("ai-stream") as HTMLElement | null;
        if (el) {
          el.classList.remove("skeleton");
          el.textContent = "AI unavailable: " + msg;
        }
      },
    );
  });
}
