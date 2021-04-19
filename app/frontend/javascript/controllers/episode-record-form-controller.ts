import { EventDispatcher } from '../utils/event-dispatcher';
import RemoteFormController from './remote-form-controller'

export default class extends RemoteFormController {
  handleSuccess(_event: any) {
    this.reloadList()
  }

  reloadList() {
    new EventDispatcher('reloadable-frame-my-episode-record-list:reload').dispatch();
  }
}
