# frozen_string_literal: true

def development?
  ENV.fetch("RAILS_ENV") == "development"
end

workers(ENV.fetch("WEB_CONCURRENCY") { 2 }.to_i) unless development?

threads_count = ENV.fetch("RAILS_MAX_THREADS") { 5 }.to_i
threads threads_count, threads_count

preload_app! unless development?

rackup DefaultRackup
port ENV.fetch("PORT") { 3000 }
environment ENV.fetch("RAILS_ENV") { "development" }

on_worker_boot do
  # Worker specific setup for Rails 4.1+
  # See: https://devcenter.heroku.com/articles/deploying-rails-applications-with-the-puma-web-server#on-worker-boot
  ActiveRecord::Base.establish_connection unless development?
end

before_fork do
  require "puma_worker_killer"

  PumaWorkerKiller.enable_rolling_restart(12 * 3600) # every 12 hours
end

# Allow puma to be restarted by `rails restart` command.
plugin :tmp_restart
