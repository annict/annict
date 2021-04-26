import axios, { AxiosResponse } from 'axios';
import { Controller } from 'stimulus';

export default abstract class extends Controller {
  static targets = ['swap'];
  static values = {
    url: String,
  }

  abstract swapSelectors: string[]

  swapTargets!: HTMLElement[];
  urlValue!: string;

  connect() {
    axios
      .get(this.urlValue)
      .then((res: AxiosResponse) => {
        this.replace(res);
      });
  }

  replace(res: AxiosResponse) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(res.data, 'text/html')

    this.swapSelectors.forEach(selector => {
      const elm = doc.querySelector(selector)
      const targetElm = this.swapTargets.find(t => t.matches(selector))

      if (elm && targetElm) {
        targetElm.innerHTML = elm.innerHTML
      }
    })
  }
}
