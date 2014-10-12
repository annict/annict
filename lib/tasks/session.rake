namespace :session do
  task sweep: :environment do
    Session.delete_all("updated_at < '#{10.days.ago.to_s(:db)}'")
  end
end
