const CHAR_MAP = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
};

export default text => text.replace(/[&<>]/g, char => CHAR_MAP[char]);
