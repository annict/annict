namespace :session do
  task sweep: :environment do
    Session.delete_all("updated_at < '#{7.days.ago.to_s(:db)}'")
  end
end
