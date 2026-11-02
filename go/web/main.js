import "basecoat-css/all";

import { initializeCopyableInputs } from "./copyable-input";
import { initializeSidebarToggle } from "./sidebar-toggle";
import { initializeSeasonFilter } from "./season-filter";

window.disableSubmitButtons = function (form) {
  form.querySelectorAll("button[type=submit]").forEach((b) => (b.disabled = true));
};

document.addEventListener("DOMContentLoaded", () => {
  initializeSidebarToggle();
  initializeSeasonFilter();
  initializeCopyableInputs();
});

console.log("Annict Go initialized");
