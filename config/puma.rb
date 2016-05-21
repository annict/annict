# frozen_string_literal: true

workers ENV.fetch("WEB_CONCURRENCY") { 2 }.to_i
threads_count = ENV.fetch("MAX_THREADS") { 5 }.to_i
threads threads_count, threads_count

preload_app!

rackup DefaultRackup
port ENV.fetch("PORT") { 3000 }
environment ENV.fetch("RAILS_ENV") { "development" }

on_worker_boot do
  # Worker specific setup for Rails 4.1+
  # See: https://devcenter.heroku.com/articles/deploying-rails-applications-with-the-puma-web-server#on-worker-boot
  ActiveRecord::Base.establish_connection
end

# Allow puma to be restarted by `rails restart` command.
plugin :tmp_restart
