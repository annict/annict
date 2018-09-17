export interface ContextData {
  readonly CSRF_PARAM: string
  readonly CSRF_TOKEN: string
  readonly ENCODED_USER_ID: string
  readonly IS_SIGNED_IN: boolean
  readonly USER_TYPE: string
  readonly VIEWER_UUID: string
}
