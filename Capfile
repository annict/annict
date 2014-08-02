# Load DSL and Setup Up Stages
require 'capistrano/setup'

# Includes default deployment tasks
require 'capistrano/deploy'

# Includes tasks from other gems included in your Gemfile
require 'capistrano/rvm'
require 'capistrano/rails'
require 'capistrano/sidekiq'
require 'capistrano/unicorn_nginx'
require 'whenever/capistrano'

# Loads custom tasks from `lib/capistrano/tasks' if you have any defined.
Dir.glob('lib/capistrano/tasks/*.cap').each { |r| import r }