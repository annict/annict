import $ from 'jquery';
import axios from 'axios';
import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static targets = ['count'];
  static values = { initIsLiked: Boolean };

  countTarget!: HTMLElement;
  initIsLikedValue!: boolean;
  isLiked!: boolean;
  isLoading!: boolean;
  likesCount!: number;
  pageCategory!: string | null;
  resourceId!: number | null;
  resourceName!: string | null;

  initialize() {
    this.resourceName = this.data.get('resourceName');
    this.resourceId = Number(this.data.get('resourceId'));
    this.pageCategory = this.data.get('pageCategory');
    this.isLiked = this.initIsLikedValue;
    this.likesCount = Number(this.countTarget.innerText);

    this.render();

    document.addEventListener('component-value-fetcher:like-button:fetched', (event: any) => {
      const likes = event.detail;
      this.likesCount = Number(this.countTarget.innerText);
      const like = likes.filter((like: { recipient_type: string; recipient_id: number }) => {
        return like.recipient_type === this.resourceName && like.recipient_id === this.resourceId;
      })[0];

      this.isLiked = !!like;
      this.render();
    });
  }

  render() {
    const iconElm = this.element.querySelector('.c-like-button__icon');

    if (!iconElm) {
      return;
    }

    if (this.isLiked) {
      this.element.classList.add('is-liked');
      iconElm.outerHTML = '<i class="c-like-button__icon fas fa-heart"></i>'; // Using outerHTML to render fontawesome icon after refetch
      this.countTarget.innerText = this.likesCount.toString();
    } else {
      this.element.classList.remove('is-liked');
      iconElm.outerHTML = '<i class="c-like-button__icon far fa-heart"></i>';
      this.countTarget.innerText = this.likesCount.toString();
    }
  }

  toggle() {
    if (this.isLoading) {
      return;
    }

    this.isLoading = true;

    if (this.isLiked) {
      axios
        .post('/api/internal/unlikes', {
          recipient_type: this.resourceName,
          recipient_id: this.resourceId,
        })
        .then(() => {
          this.isLoading = false;
          this.likesCount += -1;
          this.isLiked = false;
          this.render();
        });
    } else {
      axios
        .post('/api/internal/likes', {
          recipient_type: this.resourceName,
          recipient_id: this.resourceId,
          page_category: this.pageCategory,
        })
        .then(() => {
          this.likesCount += 1;
          this.isLiked = true;
        })
        .catch(() => {
          ($('.c-sign-up-modal') as any).modal('show');
        })
        .then(() => {
          this.isLoading = false;
          this.render();
        });
    }
  }
}
