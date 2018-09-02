type baseData = {
  csrfParam: string,
  csrfToken: string,
  domain: string,
  env: string,
  locale: string,
  isSignedIn: bool,
};

let parseBaseDataJson = json: baseData =>
  Json.Decode.{
    csrfParam: json |> field("csrfParam", string),
    csrfToken: json |> field("csrfToken", string),
    domain: json |> field("domain", string),
    env: json |> field("env", string),
    locale: json |> field("locale", string),
    isSignedIn: json |> field("isSignedIn", bool),
  };

let fetch = () =>
  Js.Promise.(
    Fetch.fetch("/api/internal/v3/base_data")
    |> then_(Fetch.Response.json)
    |> then_(json => parseBaseDataJson(json) |> resolve)
  );
