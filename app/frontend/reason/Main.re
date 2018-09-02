let listener = () => {
  let appElm = Document.getElementById("ann-app");
  let pageComponentName = Element.getAttribute(appElm, "data-component-name");
  ReactDOMRe.render(<App pageComponentName />, appElm);
};

Document.addEventListener("turbolinks:load", listener);

Turbolinks.start();
