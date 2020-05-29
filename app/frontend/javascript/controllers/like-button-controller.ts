import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

export default class extends Controller {
  static targets = ['count'];

  countTarget!: HTMLElement;
  resourceName!: string | null;
  resourceId!: number | null;
  likesCount!: number;
  pageCategory!: string | null;
  isLiked!: boolean;
  isLoading!: boolean;

  initialize() {
    this.resourceName = this.data.get('resourceName');
    this.resourceId = Number(this.data.get('resourceId'));
    this.pageCategory = this.data.get('pageCategory');

    document.addEventListener('user-data-fetcher:likes:fetched', ({ detail: { likes } }: any) => {
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
        .post('/api/internal/likes/unlike', {
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
