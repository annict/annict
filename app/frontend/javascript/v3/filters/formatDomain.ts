export default value => {
  if (!value) {
    return ''
  }

  return new URL(value).hostname
}
