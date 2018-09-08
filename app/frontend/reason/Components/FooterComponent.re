let component = ReasonReact.statelessComponent("Footer");

let make = _children => {
  ...component,
  render: _self => <div> {ReasonReact.string("footer")} </div>,
};
