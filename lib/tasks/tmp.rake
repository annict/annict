namespace :tmp do
  task update_next_episode_id: :environment do
    Work.find_each do |work|
      puts work.title
      prev_episode = nil

      work.episodes.order(:sort_number).find_each do |episode|
        if prev_episode.present?
          prev_episode.update_column(:next_episode_id, episode.id)
        end

        prev_episode = episode
      end
    end
  end

  task build_setting: :environment do
    User.all.order(:id).each do |user|
      puts "user_id: #{user.id}"
      user.build_setting
      user.save
    end
  end
end
