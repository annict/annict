namespace :counter_cache do
  task refresh_watchers_count: :environment do
    Work.find_each do |work|
      watchers_count = work.latest_statuses.with_kind(:wanna_watch, :watching, :watched).count
      work.update_column(:watchers_count, watchers_count)

      puts "Work ID: #{work.id} のwatchers_countを更新しました。"
    end
  end

  task refresh_items_count_on_works: :environment do
    Work.find_each do |work|
      puts "work_id: #{work.id}"
      Work.reset_counters(work.id, :items)
    end
  end
end
