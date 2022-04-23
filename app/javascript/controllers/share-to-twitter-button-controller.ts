import { Controller } from '@hotwired/stimulus';

import urlParams from '../utils/url-params';

export default class extends Controller {
  text!: string | null;
  url!: string | null;
  hashtags!: string | null;

  initialize() {
    this.text = this.data.get('text');
    this.url = this.data.get('url');
    this.hashtags = this.data.get('hashtags');
  }

  get baseTweetUrl() {
    return 'https://twitter.com/intent/tweet';
  }

  get tweetUrl() {
    const params = urlParams({
      text: `${this.text} | Annict`,
      url: this.url,
      hashtags: this.hashtags || '',
    });

    return `${this.baseTweetUrl}?${params}`;
  }

  open() {
    const left = (screen.width - 640) / 2;
    const top = (screen.height - 480) / 2;
    return open(this.tweetUrl, '', `width=640,height=480,left=${left},top=${top}`);
  }
}
