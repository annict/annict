namespace :tmp do
  task remove_hash_from_twitter_hashtag: :environment do
    Work.find_each do |work|
      if work.twitter_hashtag.present?
        puts work.title
        work.update_column(:twitter_hashtag, work.twitter_hashtag.sub(/^#/, ""))
      end
    end
  end
end
