import RemoteFormController from './remote-form-controller'

export default class extends RemoteFormController {
  handleSuccess(_event: any) {
    console.log('success!')
  }
}
