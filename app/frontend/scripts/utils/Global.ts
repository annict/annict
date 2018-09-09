import { Ann } from './Ann'

declare global {
  interface Window {
    readonly ann: Ann
    dataLayer: any[]
  }
}
