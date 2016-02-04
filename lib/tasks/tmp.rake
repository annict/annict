namespace :tmp do
  task fix_watched_episode_ids: :environment do
    User.find_each do |user|
      user.latest_statuses.find_each do |latest_status|
        episode_ids = user.checkins.where(work: latest_status.work).pluck(:episode_id)
        if (latest_status.watched_episode_ids - episode_ids).present?
          puts "----------------------"
          puts "user: #{user.id}"
          puts "work: #{latest_status.work.id}"
          puts "episode_ids: #{episode_ids.join(', ')}"
          puts "latest_status.watched_episode_ids: #{latest_status.watched_episode_ids.join(', ')}"
          latest_status.watched_episode_ids = latest_status.watched_episode_ids & episode_ids
          latest_status.save!
        end
      end
    end
  end
end
