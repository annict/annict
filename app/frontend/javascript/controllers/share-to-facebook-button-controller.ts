import { Controller } from '@hotwired/stimulus';

import urlParams from '../utils/url-params';

export default class extends Controller {
  url!: string | null;
  appId!: string | null;

  initialize() {
    this.url = this.data.get('url');
    this.appId = this.data.get('appId');
  }

  get baseShareUrl() {
    return 'https://www.facebook.com/sharer/sharer.php';
  }

  get shareUrl() {
    const params = urlParams({
      u: this.url,
      display: 'popup',
      ref: 'plugin',
      src: 'like',
      kid_directed_site: 0,
      app_id: this.appId,
    });

    return `${this.baseShareUrl}?${params}`;
  }

  open() {
    const left = (screen.width - 640) / 2;
    const top = (screen.height - 480) / 2;
    return open(this.shareUrl, '', `width=640,height=480,left=${left},top=${top}`);
  }
}
