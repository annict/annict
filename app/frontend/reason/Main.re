let listener = () => {
  let elm = Document.getElementById("ann-react-app");
  let componentName = Element.getAttribute(elm, "data-component-name");
  ReactDOMRe.render(<App componentName />, elm);
};

Document.addEventListener("turbolinks:load", listener);

Turbolinks.start();
