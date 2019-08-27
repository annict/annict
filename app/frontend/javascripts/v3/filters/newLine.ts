export default value => {
  if (!value) {
    return ''
  }

  return value.replace(/\n{3,}/g, '<br><br>').replace(/\n/g, '<br>')
}
