type state = {
  loading: bool,
  baseData: BaseData.baseData,
};

type action =
  | FetchBaseData
  | LoadBaseData(BaseData.baseData);

let appComponent = ReasonReact.reducerComponent("App");

let make = (~pageComponentName, _children) => {
  ...appComponent,
  initialState: () => {
    loading: false,
    baseData: {
      csrfParam: "",
      csrfToken: "",
      domain: "",
      env: "",
      locale: "",
      isSignedIn: false,
    },
  },
  reducer: (action, state) =>
    switch (action) {
    | FetchBaseData =>
      Js.log(
        "FetchBaseData called. state.loading: "
        ++ string_of_bool(state.loading),
      );
      ReasonReact.UpdateWithSideEffects(
        {...state, loading: true},
        (
          self => {
            Js.log(
              "UpdateWithSideEffects called. state.loading: "
              ++ string_of_bool(self.state.loading),
            );
            Js.Promise.(
              BaseData.fetch()
              |> then_(baseData =>
                   self.send(LoadBaseData(baseData)) |> resolve
                 )
              |> ignore
            );
          }
        ),
      );
    | LoadBaseData(baseData) =>
      ReasonReact.Update({loading: false, baseData})
    },
  didMount: self => self.send(FetchBaseData),
  render: _self =>
    switch (pageComponentName) {
    | "Home" => <Components.Home />
    | _ => <Components.Blank />
    },
};
