import "basecoat-css/all";

import { initializeSidebarToggle } from "./sidebar-toggle";

window.disableSubmitButtons = function (form) {
  form.querySelectorAll("button[type=submit]").forEach((b) => (b.disabled = true));
};

document.addEventListener("DOMContentLoaded", () => {
  initializeSidebarToggle();
});

console.log("Annict Go initialized");
