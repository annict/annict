import moment from 'moment'

export default value => {
  if (!value) {
    return ''
  }

  return moment(String(value)).format('YYYY/MM/DD HH:mm:ss')
}
