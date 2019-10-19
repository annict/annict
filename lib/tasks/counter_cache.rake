namespace :counter_cache do
  task refresh: :environment do
    %w(
      refresh_watchers_count
      refresh_notifications_count
      refresh_record_comments_count
      refresh_work_records_with_body_count
      refresh_records_count
    ).each do |task_name|
      puts "============== #{task_name} =============="
      Rake::Task["counter_cache:#{task_name}"].invoke
    end
  end

  task refresh_watchers_count: :environment do
    Work.find_each do |work|
      kinds = %w(wanna_watch watching watched)
      watchers_count = work.latest_statuses.with_kind(*kinds).count

      if work.watchers_count != watchers_count
        work.update_column(:watchers_count, watchers_count)
        puts "Work ID: #{work.id} の watchers_count を更新しました。"
      end
    end
  end

  task refresh_notifications_count: :environment do
    User.find_each do |user|
      user.update_column(:notifications_count, user.notifications.unread.count)
      puts "User ID: #{user.id} の notifications_count を更新しました。"
    end
  end

  task refresh_record_comments_count: :environment do
    Episode.published.find_each do |episode|
      record_comments_count = episode.episode_records.with_body.count
      if record_comments_count != episode.episode_record_bodies_count
        episode.update_column(:episode_record_bodies_count, record_comments_count)
        puts "Episode ID: #{episode.id} - #{record_comments_count}"
      else
        puts "Episode ID: #{episode.id} - skipped."
      end
    end
  end

  task refresh_work_records_with_body_count: :environment do
    Work.published.find_each do |work|
      work_records_with_body_count = work.work_records.published.with_body.count
      if work_records_with_body_count != work.work_records_with_body_count
        work.update_column(:work_records_with_body_count, work_records_with_body_count)
        puts "Work ID: #{work.id} - #{work_records_with_body_count}"
      else
        puts "Work ID: #{work.id} - skipped."
      end
    end
  end

  task refresh_records_count: :environment do
    # Attach monky patch to update readonly field
    [User, Work].each do |model|
      model.class_eval do
        def self.readonly_attributes
          []
        end
      end
    end

    ActiveRecord::Base.transaction do
      User.find_each do |u|
        puts "User #{u.id}"
        u.update_column(:records_count, u.records.published.count)
      end
    end

    ActiveRecord::Base.transaction do
      Work.published.find_each do |w|
        puts "Work #{w.id}"
        w.update_column(:records_count, w.records.published.count)
      end
    end
  end
end
