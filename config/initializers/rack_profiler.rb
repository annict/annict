# frozen_string_literal: true

if Rails.env.development?
  require "rack-mini-profiler"

  Rack::MiniProfiler.config.position = "bottom-right"

  # initialization is skipped so trigger it
  Rack::MiniProfilerRails.initialize!(Rails.application)
end
