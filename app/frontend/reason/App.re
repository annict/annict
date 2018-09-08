type state = {
  loading: bool,
  baseData: BaseData.baseData,
  pageData: PageData.pageData,
};

type action =
  | FetchBaseData
  | LoadBaseData(BaseData.baseData)
  | SetupAnalytics(BaseData.baseData);

let appComponent = ReasonReact.reducerComponent("App");

let make = (~pageComponentName, ~pageCategory, _children) => {
  ...appComponent,
  initialState: () => {
    loading: false,
    baseData: {
      viewerUUID: "",
      csrfParam: "",
      csrfToken: "",
      domain: "",
      encodedUserId: "",
      env: "",
      gaTrackingId: "",
      isSignedIn: false,
      locale: "",
      userType: "",
    },
    pageData: {
      pageCategory: pageCategory,
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
              "FetchBaseData - UpdateWithSideEffects called. state.loading: "
              ++ string_of_bool(self.state.loading),
            );
            Js.Promise.(
              BaseData.fetch()
              |> then_(baseData =>
                   {
                     self.send(LoadBaseData(baseData));
                     self.send(SetupAnalytics(baseData));
                   }
                   |> resolve
                 )
              |> ignore
            );
          }
        ),
      );
    | SetupAnalytics(baseData) =>
      ReasonReact.SideEffects(
        (
          self => {
            let load = [%raw
              {|
              function(gaTrackingId, viewerUUID, encodedUserId, userType, pageCategory) {
                window.dataLayer = window.dataLayer || [];
                function gtag(){dataLayer.push(arguments);}
                gtag("js", new Date());

                gtag("config", gaTrackingId, {
                  "client_id": viewerUUID,
                  "user_id": encodedUserId,
                  "dimension1": userType,
                  "dimension2": pageCategory,
                  "custom_map": {
                    "dimension1": userType,
                    "dimension2": pageCategory,
                  }
                });
              }
              |}
            ];
            load(
              baseData.gaTrackingId,
              baseData.viewerUUID,
              baseData.encodedUserId,
              baseData.userType,
              self.state.pageData.pageCategory,
            );
          }
        ),
      )
    | LoadBaseData(baseData) =>
      ReasonReact.Update({loading: false, baseData, pageData: state.pageData})
    },
  didMount: self => self.send(FetchBaseData),
  render: self => {
    let pageComponent =
      switch (pageComponentName) {
      | "Home" => <PageComponents.Home baseData={self.state.baseData} />
      | _ => <PageComponents.Blank />
      };
    <> pageComponent </>;
  },
};
