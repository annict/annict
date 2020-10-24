import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';
import lazyLoad from '../utils/lazy-load';

export default class extends Controller {
  static targets = ['nextButton'];

  readonly nextButtonTarget!: Element;

  activityGroupId!: string | null;
  cursor!: string | null;

  initialize() {
    this.activityGroupId = this.data.get('activityGroupId');
    this.cursor = this.data.get('cursor');
  }

  next() {
    axios
      .get('/api/internal/activities', {
        params: {
          activity_group_id: this.activityGroupId,
          cursor: this.cursor,
        },
      })
      .then((res) => {
        const activityCardsElm = this.element.querySelector('.c-timeline__activity-cards');

        if (activityCardsElm) {
          this.nextButtonTarget.classList.add('d-none');
          activityCardsElm.innerHTML += res.data;
          new EventDispatcher('user-data-fetcher:refetch').dispatch();
          lazyLoad.update();
        }
      });
  }
}
