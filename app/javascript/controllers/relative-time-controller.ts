import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { Controller } from '@hotwired/stimulus';

dayjs.extend(relativeTime);

export default class extends Controller {
  connect() {
    this.element.textContent = this.relativeTime;
    this.element.setAttribute('title', this.absoluteTime);
  }

  get time() {
    return dayjs(this.data.get('time')!);
  }

  get absoluteTime() {
    return dayjs(this.time).format('YYYY-MM-DD HH:mm');
  }

  get relativeTime() {
    const currentTime = dayjs();

    const passageDays = dayjs(currentTime).diff(this.time, 'day');

    if (passageDays > 3) {
      return this.time.format('YYYY-MM-DD');
    } else {
      return this.time.fromNow();
    }
  }
}
