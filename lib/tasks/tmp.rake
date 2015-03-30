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

  task update_checks: :environment do
    User.find_each do |user|
      puts "user: #{user.id}"
      watching_works = user.works.watching
      user.checks.each do |check|
        check.destroy unless watching_works.exists?(id: check.work.id)
      end
    end
  end

  task convert_image_for_paperclip: :environment do
    host = "http://d3a8d1smk6xli.cloudfront.net/"

    Item.where(tombo_image_file_name: nil).order(:id).each do |i|
      puts "Item: #{i.id}"
      image_url = host + i.image_uid
      i.tombo_image = URI.parse(image_url)
      i.save
    end

    Profile.where(tombo_avatar_file_name: nil).where.not(avatar_uid: nil).order(:id).each do |p|
      puts "avatar Profile: #{p.id}"
      image_url = host + p.avatar_uid
      p.tombo_avatar = URI.parse(image_url)
      p.save(validate: false)
    end

    Profile.where(tombo_background_image_file_name: nil).where.not(background_image_uid: nil).order(:id).each do |p|
      puts "background Profile: #{p.id}"
      image_url = host + p.background_image_uid
      p.tombo_background_image = URI.parse(image_url)
      p.save(validate: false)
    end
  end
end
