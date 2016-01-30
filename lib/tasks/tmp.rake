namespace :tmp do
  task create_latest_status: :environment do
    Status.where(latest: true).find_each do |s|
      puts "status: #{s.id}"
      latest_status = s.user.latest_statuses.find_or_initialize_by(work: s.work)
      latest_status.kind = s.kind
      latest_status.save!
    end
  end

  task move_checks_to_latest_statuses: :environment do
    Check.find_each do |c|
      puts "check: #{c.id}"

      episode_ids = c.work.episodes.order(:sort_number).pluck(:id)
      if c.episode_id.present?
        begin
          index = episode_ids.index(c.episode_id) - 1
          watched_episode_ids = index > 0 ? episode_ids[0..index] : []
        rescue
          binding.pry
        end
      else
        watched_episode_ids = episode_ids
      end
      watched_episode_ids = (watched_episode_ids + c.skipped_episode_ids).uniq

      begin
        latest_status = c.user.latest_statuses.find_by(work: c.work)
        if latest_status.blank?
          status = c.user.statuses.where(work: c.work).order(id: :desc).first
          latest_status = c.user.latest_statuses.create(work: c.work, kind: status.kind)
        end
        latest_status.watched_episode_ids = watched_episode_ids
        latest_status.position = c.position
        latest_status.save!
      rescue
        binding.pry
      end
    end
  end
end
