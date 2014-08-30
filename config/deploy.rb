lock '3.2.1'

set :application, 'annict'
set :repo_url,    'git@github.com:annict/annict.git'

set :deploy_to,    '/home/annict'
set :linked_files, %w(config/application.yml config/database.yml)
set :linked_dirs,  %w(log)

set :rbenv_type, :system
set :rbenv_ruby, '2.1.2'

# set :sidekiq_pid, File.join(shared_path, 'pids', 'sidekiq.pid')
# set :sidekiq_options, "-C #{current_path}/config/sidekiq.yml"
