import { Controller } from 'stimulus';

export default class extends Controller {
  static targets = ['frame'];

  boundReload!: any;
  // `turbo-frame` 要素を参照するためのTarget
  // NOTE:
  //   `this.element` で参照できるが、`this.element` の型は `Element` になっており、
  //   `src` 属性を参照するために `this.element.src` などと書くとTypeScriptの型エラーになるためこうしている
  //   あと本来は `HTMLIFrameElement` ではなく `FrameElement` だが、`FrameElement` がインポートできない気がするのでこうしている
  frameTarget!: HTMLIFrameElement;

  initialize() {
    this.boundReload = this.reload.bind(this);
  }

  connect() {
    document.addEventListener('trackable-episode-list:reload', this.boundReload);
  }

  disconnect() {
    document.removeEventListener('trackable-episode-list:reload', this.boundReload);
  }

  reload() {
    // https://github.com/hotwired/turbo/pull/206 がマージされたら `this.element.reload()` と書けそう
    const { src } = this.frameTarget
    this.frameTarget.src = ''
    this.frameTarget.src = src
  }
}
