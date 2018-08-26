[@bs.val] [@bs.scope "document"]
external addEventListener: (string, unit => unit) => unit = "";

module Turbolinks = {
  [@bs.module "turbolinks"] external start: unit => unit = "";
};

let listener = () => {
  ReactDOMRe.renderToElementWithId(<Component1 message="Hello!" />, "index1");
  ReactDOMRe.renderToElementWithId(
    <Component2 greeting="Hello!" />,
    "index2",
  );
};

addEventListener("turbolinks:load", listener);

Turbolinks.start();
