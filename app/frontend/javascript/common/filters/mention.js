export default val => {
  if (!val) {
    return '';
  }

  const pattern = /@[A-Za-z0-9_]+/g;
  const matches = val.match(pattern);

  if (matches) {
    matches.forEach(match => {
      const url = `<a href="${gon.annict.url}/${match}">${match}</a>`;
      val = val.replace(match, url);
    });
  }

  return val;
};
