import { EventDispatcher } from '../utils/event-dispatcher';
import FrameFormController from './frame-form-controller'

export default class extends FrameFormController {
  static targets = [];

  async handleSuccess(event: any) {
    super.handleSuccess(event)
    this.reloadList()
  }

  reloadList() {
    new EventDispatcher('reloadable--episode-record-list-group:reload').dispatch();
  }
}
