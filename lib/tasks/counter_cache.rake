namespace :counter_cache do
  task refresh_watchers_count: :environment do
    Work.find_each do |work|
      watchers_count = work.latest_statuses.with_kind(:wanna_watch, :watching, :watched).count
      work.update_column(:watchers_count, watchers_count)

      puts "Work ID: #{work.id} の watchers_count を更新しました。"
    end
  end

  task refresh_notifications_count: :environment do
    User.find_each do |user|
      user.update_column(:notifications_count, user.notifications.unread.count)
      puts "User ID: #{user.id} の notifications_count を更新しました。"
    end
  end
end
