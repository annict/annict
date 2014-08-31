server 'annict3-web', roles: %w(web app db)

set :branch, 'master'

set :ssh_options, {
  keys:          %w(~/.ssh/bojovs/annict/annict),
  forward_agent: true,
  user:          'annict'
}
