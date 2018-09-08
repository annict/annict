type baseData = {
  viewerUUID: string,
  csrfParam: string,
  csrfToken: string,
  domain: string,
  encodedUserId: string,
  env: string,
  gaTrackingId: string,
  isSignedIn: bool,
  locale: string,
  userType: string,
};

let parseBaseDataJson = json: baseData =>
  Json.Decode.{
    viewerUUID: json |> field("viewerUUID", string),
    csrfParam: json |> field("csrfParam", string),
    csrfToken: json |> field("csrfToken", string),
    domain: json |> field("domain", string),
    encodedUserId: json |> field("encodedUserId", string),
    env: json |> field("env", string),
    gaTrackingId: json |> field("gaTrackingId", string),
    isSignedIn: json |> field("isSignedIn", bool),
    locale: json |> field("locale", string),
    userType: json |> field("userType", string),
  };

let fetch = () =>
  Js.Promise.(
    Fetch.fetch("/api/internal/v3/base_data")
    |> then_(Fetch.Response.json)
    |> then_(json => parseBaseDataJson(json) |> resolve)
  );
