let component = ReasonReact.statelessComponent("Home");

let make = (~baseData, _children) => {
  ...component,
  render: _self => <> <Components.Footer /> <Components.Footer /> </>,
};
