module.exports = function (api) {
  return {
    presets: [
      [
        '@babel/preset-env',
        {
          targets: '> 0.5%, last 2 versions, Firefox ESR, not IE 11, not IE_Mob 11, not dead',
          useBuiltIns: 'usage',
          corejs: 3,
          forceAllTransforms: api.env('production'),
        },
        '@babel/preset-typescript',
      ],
    ],

    plugins: [
      '@babel/plugin-transform-typescript',
      '@babel/plugin-proposal-class-properties',
      '@babel/plugin-transform-runtime',
      '@babel/plugin-proposal-object-rest-spread',
    ],
  };
};
