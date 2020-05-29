import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';
import lazyLoad from "../utils/lazy-load";

export default class extends Controller {
  static targets = ['nextButton'];

  readonly nextButtonTarget!: Element;

  username!: string | null;
  pageCategory!: string | null;
  cursor!: string | null;

  initialize() {
    this.username = this.data.get('username');
    this.pageCategory = this.data.get('pageCategory');
    this.cursor = this.data.get('cursor');
  }

  next() {
    axios
      .get('/api/internal/activity_groups', {
        params: {
          username: this.username,
          page_category: this.pageCategory,
          cursor: this.cursor,
        },
      })
      .then((res) => {
        const activitiesElm = this.element.querySelector('.c-timeline__activities');

        if (activitiesElm) {
          this.nextButtonTarget.classList.add('d-none');
          activitiesElm.innerHTML += res.data;
          new EventDispatcher('user-data-fetcher:refetch').dispatch();
          lazyLoad.update();
        }
      });
  }
}
