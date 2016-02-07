namespace :counter_cache do
  task refresh: :environment do
    %w(
      refresh_watchers_count
      refresh_notifications_count
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
end
