import Modal from "bootstrap/js/dist/modal";
import { Controller } from "@hotwired/stimulus";

import { EventDispatcher } from "../utils/event-dispatcher";
import fetcher from "../utils/fetcher";

export default class extends Controller {
  static classes = ["loading"];
  static values = {
    episodeId: Number,
  };

  declare readonly episodeIdValue: number;
  declare readonly loadingClass: string;
  declare isLoading: boolean;

  startLoading() {
    this.isLoading = true;
    this.element.classList.add(this.loadingClass);
    this.element.setAttribute("disabled", "true");
  }

  endLoading() {
    this.isLoading = false;
    this.element.classList.remove(this.loadingClass);
  }

  reloadList() {
    new EventDispatcher("reloadable--trackable-episode-list:reload").dispatch();
    new EventDispatcher("reloadable--tracking-offcanvas:reload").dispatch();
  }

  skip() {
    if (this.isLoading) {
      return;
    }

    this.startLoading();

    fetcher
      .post("/api/internal/skipped_episodes", {
        episode_id: this.episodeIdValue,
      })
      .then(() => {
        this.endLoading();
        this.reloadList();
      })
      .catch(() => {
        new Modal(".c-sign-up-modal").show();
      });
  }
}
