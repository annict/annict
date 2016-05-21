# frozen_string_literal: true

Rails.application.config.assets.precompile += %w(
  db/application.scss
  db/application.js.coffee
  v1/application_mobile.js.coffee
  v1/application.js.coffee
  v1/application_mobile.scss
  v1/application.scss
  v2/mobile.js.coffee
  v2/pc.js.coffee
  v2/mobile.scss
  v2/pc.scss
  v3/mobile.js.coffee
  v3/pc.js.coffee
  v3/base.scss
  v3/mobile.scss
  v3/pc.scss
)
