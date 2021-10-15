import { EventDispatcher } from '../../utils/event-dispatcher';
import BasicFormController from './basic-form-controller';

export default class extends BasicFormController {
  static targets = [];

  async handleSuccess(event: any) {
    super.handleSuccess(event);
    new EventDispatcher('record-rating:reset').dispatch();
    this.reloadList();
  }

  reloadList() {
    new EventDispatcher('reloadable--episode-record-list-group:reload').dispatch();
    new EventDispatcher('reloadable--record-list:reload').dispatch();
    new EventDispatcher('reloadable--record:reload').dispatch();
    new EventDispatcher('reloadable--trackable-episode-list:reload').dispatch();
  }
}
