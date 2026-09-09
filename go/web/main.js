import "basecoat-css/all";

import { initializeSidebarToggle } from "./sidebar-toggle";
import { initializeSeasonFilter } from "./season-filter";

window.disableSubmitButtons = function (form) {
  form.querySelectorAll("button[type=submit]").forEach((b) => (b.disabled = true));
};

document.addEventListener("DOMContentLoaded", () => {
  initializeSidebarToggle();
  initializeSeasonFilter();
});

console.log("Annict Go initialized");
