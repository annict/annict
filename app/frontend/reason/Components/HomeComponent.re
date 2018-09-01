let component = ReasonReact.statelessComponent("Home");

let make = _children => {
  ...component,
  render: _self => <div> {ReasonReact.string("Hello Home!")} </div>,
};
