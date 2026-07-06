import { Summary } from "../../types";
import { formatCapturedAt } from "../format";
import { Snapshot } from "../types";
import { requiredEl } from "../ui";

export function renderStatus(summary: Summary, previousSnapshot: Snapshot | null): void {
  const hotspots = summary.hotspots.length;
  if (summary.moduleCount === 0) {
    requiredEl<HTMLElement>("status").textContent =
      "No modules were identified. Run from the project root or use --lang all.";
    return;
  }

  requiredEl<HTMLElement>("status").textContent =
    `${summary.moduleCount} modules monitored, ${hotspots} active hotspot${hotspots === 1 ? "" : "s"}.`;

  const historyBadge = requiredEl<HTMLElement>("history-badge");
  historyBadge.textContent = previousSnapshot
    ? `Comparing against ${formatCapturedAt(previousSnapshot.capturedAt)}`
    : "No previous baseline";
}
