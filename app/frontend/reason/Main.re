let listener = () => {
  let appElm = Document.getElementById("ann-app");
  let pageComponentName = Element.getAttribute(appElm, "data-component-name");
  let pageCategory = Element.getAttribute(appElm, "data-page-category");
  ReactDOMRe.render(<App pageComponentName pageCategory />, appElm);
};

Document.addEventListener("turbolinks:load", listener);

Turbolinks.start();
