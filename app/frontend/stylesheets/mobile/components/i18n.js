export default function({ app, env, isHMR, req, store }) {
  // If middleware is called from hot module replacement, ignore it
  if (isHMR) return;

  const headers = req && req.headers ? Object.assign({}, req.headers) : {};
  const domain = headers.host ? headers.host.replace(/:[0-9]+$/, '') : '';
  const locale = domain === env.jpDomain ? 'ja' : 'en';

  store.commit('setLocale', locale);

  app.i18n.locale = store.state.locale;
}
