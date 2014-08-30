lock '3.2.1'

set :application, 'annict'
set :repo_url,    'git@github.com:annict/annict.git'

set :deploy_to,    '/home/annict'
set :linked_files, %w(config/application.yml config/database.yml)
set :linked_dirs,  %w(log)

set :rbenv_type, :system
set :rbenv_ruby, '2.1.2'

after 'deploy:publishing', 'eye:unicorn:restart'
after 'deploy:publishing', 'eye:sidekiq:restart'
