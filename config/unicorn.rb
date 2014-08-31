application  = 'annict'
app_path     = "/home/#{application}"
current_path = "#{app_path}/current"
shared_path  = "#{app_path}/shared"

listen "#{shared_path}/sockets/unicorn.sock"
pid    "#{shared_path}/pids/unicorn.pid"

worker_processes 5
# ダウンタイムをなくす
preload_app true

# Capistrano 用に RAILS_ROOT を指定
working_directory current_path

if 'production' == ENV['RAILS_ENV']
  stderr_path = "#{shared_path}/log/unicorn.stderr.log"
  stdout_path = "#{shared_path}/log/unicorn.stdout.log"
end

# ログ
stderr_path File.expand_path('log/unicorn.log', ENV['RAILS_ROOT'])
stdout_path File.expand_path('log/unicorn.log', ENV['RAILS_ROOT'])

before_fork do |server, worker|
  defined?(ActiveRecord::Base) and ActiveRecord::Base.connection.disconnect!

  old_pid = "#{server.config[:pid]}.oldbin"
  if old_pid != server.pid
    begin
      Process.kill :QUIT, File.read(old_pid).to_i
    rescue Errno::ENOENT, Errno::ESRCH
    end
  end
end

after_fork do |server, worker|
  defined?(ActiveRecord::Base) and ActiveRecord::Base.establish_connection
end
