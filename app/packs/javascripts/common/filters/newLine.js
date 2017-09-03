export default val => {
  if (!val) {
    return '';
  }

  return val.replace(/\n{3,}/g, '<br><br>').replace(/\n/g, '<br>');
};
