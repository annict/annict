lock '3.2.1'

set :application, 'annict'
set :repo_url,    'git@github.com:bojovs/annict.git'

set :deploy_to,    '/home/annict'
set :linked_files, %w(config/application.yml config/database.yml)
set :linked_dirs,  %w(log)

set :rvm_type,         :system             # Defaults to: :auto
set :rvm_ruby_version, 'ruby-2.1.1@annict' # Defaults to: 'default'

set :sidekiq_pid, File.join(shared_path, 'pids', 'sidekiq.pid')
set :sidekiq_options, "-C #{current_path}/config/sidekiq.yml"