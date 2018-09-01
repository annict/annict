let appComponent = ReasonReact.statelessComponent("App");

let make = (~componentName, _children) => {
  ...appComponent,
  render: _self =>
    switch (componentName) {
    | "Home" => <Components.Home />
    | _ => <Components.Blank />
    },
};
