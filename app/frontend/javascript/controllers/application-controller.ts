import { Controller } from 'stimulus';

import lazyLoad from '../utils/lazy-load';
import { UserDataFetcher } from '../utils/user-data-fetcher';

export default class extends Controller {
  pageCategory!: string;
  params!: {};

  async initialize() {
    lazyLoad.update();

    this.pageCategory = this.data.get('pageCategory') || '';
    try {
      this.params = JSON.parse(this.data.get('params') || '{}');
    } catch (ex) {
      this.params = {};
    }

    await new UserDataFetcher(this.pageCategory, this.params).start();
  }
}
