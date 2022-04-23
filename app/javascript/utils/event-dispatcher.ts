export class EventDispatcher {
  event!: CustomEvent;

  constructor(eventName: string, detail: any = {}) {
    this.event = new CustomEvent(eventName, {
      detail,
    });
  }

  dispatch() {
    document.dispatchEvent(this.event);
  }
}
