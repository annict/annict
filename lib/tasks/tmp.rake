namespace :tmp do
  task remove_hash_from_twitter_hashtag: :environment do
    Work.find_each do |work|
      if work.twitter_hashtag.present?
        puts work.title
        work.update_column(:twitter_hashtag, work.twitter_hashtag.sub(/^#/, ""))
      end
    end
  end

  task move_next_episode_id_to_prev_episode_id: :environment do
    Episode.where.not(next_episode_id: nil).find_each do |e|
      puts "episode_id: #{e.id}"
      e.old_next_episode.update_column(:prev_episode_id, e.id)
    end
  end
end
