/* eslint-disable @typescript-eslint/no-explicit-any */

import { Controller } from "@hotwired/stimulus";

export default class extends Controller {
  static targets = ["value"];

  declare readonly valueTarget: HTMLElement;
  declare charactersCounterId: string;
  declare boundUpdate: any;

  initialize() {
    this.charactersCounterId = this.data.get("charactersCounterId") ?? "id";
    this.boundUpdate = this.update.bind(this);
    this.valueTarget.innerText = "0";
  }

  connect() {
    document.addEventListener("characters-counter:update", this.boundUpdate);
  }

  disconnect() {
    document.removeEventListener("characters-counter:update", this.boundUpdate);
  }

  update({ detail: { charactersCounterId, charactersCount } }: any) {
    if (this.charactersCounterId === charactersCounterId) {
      this.valueTarget.innerText = charactersCount;
    }
  }
}
