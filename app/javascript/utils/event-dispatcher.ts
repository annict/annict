/* eslint-disable @typescript-eslint/no-explicit-any */

export class EventDispatcher {
  declare event: CustomEvent;

  constructor(eventName: string, detail: any = {}) {
    this.event = new CustomEvent(eventName, {
      detail,
    });
  }

  dispatch() {
    document.dispatchEvent(this.event);
  }
}
