server 'annict-staging', roles: %w(web app db)

set :branch, 'staging'

set :ssh_options, {
  keys:          %w(~/.ssh/id_rsa),
  forward_agent: true,
  user:          'annict'
}