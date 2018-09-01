[@bs.val]
external addEventListener: (string, unit => unit) => unit =
  "document.addEventListener";
[@bs.val]
external getElementById: string => Dom.element = "document.getElementById";
