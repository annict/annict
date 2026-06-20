import "basecoat-css/all";

window.disableSubmitButtons = function (form) {
  form.querySelectorAll("button[type=submit]").forEach((b) => (b.disabled = true));
};

console.log("Annict Go initialized");
