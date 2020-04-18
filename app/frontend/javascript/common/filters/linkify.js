import * as linkify from 'linkifyjs';
import mention from 'linkifyjs/plugins/mention';

mention(linkify);

export default function(str) {
  const links = linkify.find(str);

  for (let link of Array.from(links)) {
    str = str.replace(link.value, _getTag(link));
  }

  return str;
}

var _getTag = function(link) {
  switch (link.type) {
    case 'url':
      return `<a href='${link.href}'>${link.value}</a>`;
    case 'mention':
      return `<a href='${gon.annict.url}/${link.value}'>${link.value}</a>`;
  }
};
