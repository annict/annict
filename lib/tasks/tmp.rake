namespace :tmp do
  task delete_db_activities: :environment do
    DbActivity.delete_all
  end
end
