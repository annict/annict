// declare var WebFont: object
interface Window {
  AnnConfig: AnnConfigType;
  WebFont: WebFontType;
}

interface AnnConfigType {
  facebook: { appId: string };
  isDomainJp(): boolean;
  isSignedIn(): boolean;
  locale: string;
}

interface WebFontType {
  load(obj: object): void;
}

interface JQuery {
  modal(str: string): void;
}

// https://github.com/apollographql/apollo-link/issues/1131
declare type GlobalFetch = WindowOrWorkerGlobalScope
