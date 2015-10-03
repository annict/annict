namespace :tmp do
  task update_work_id: :environment do
    Comment.find_each do |c|
      c.update_column(:work_id, c.checkin.work_id)
    end
  end
end
