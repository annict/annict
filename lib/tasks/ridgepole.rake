namespace :ridgepole do
  desc 'Run ridgepole'
  task :apply, :with_dry_run do |t, args|
    opts = []
    opts << "-c #{Rails.root.join('config', 'database.yml')}"
    opts << "-f #{Rails.root.join('Schemafile')}"
    opts << "-E #{ENV['RAILS_ENV'] || 'development'}"
    opts << '--apply'
    opts << (args[:with_dry_run] ? '--dry-run' : '')
    opts << '--enable-foreigner'

    cmd = "ridgepole #{opts.join(' ')}"
    system(cmd)
  end

  desc 'Run ridgepole dry-run'
  task 'apply:dry-run' do
    Rake::Task['ridgepole:apply'].invoke(true)
  end

  desc 'Rebuild database and run ridgepole:apply'
  task :clean do
    Rake::Task['db:drop'].invoke
    Rake::Task['db:create'].invoke
    Rake::Task['ridgepole:apply'].invoke
  end

  desc 'Run ridgepole:clean and db:seed'
  task :reset do
    Rake::Task['ridgepole:clean'].invoke
    Rake::Task['db:seed'].invoke
  end
end
