# frozen_string_literal: true

namespace :work do
  # 指定したWorkを削除する
  # コマンド実行例: rake work:hide_overlapped_work[4458,4485]
  task :hide_overlapped_work, %i[target_work_id original_work_id] => :environment do |_, args|
    # 削除対象のWork
    target_work = Anime.find(args[:target_work_id])
    # オリジナルのWork
    original_work = Anime.find(args[:original_work_id])

    ActiveRecord::Base.transaction do
      [
        {resource_class: Activity, column: :recipient}
      ].each do |hash|
        update_or_delete_pol_resource(hash[:resource_class], hash[:column], target_work, original_work)
      end

      [ChannelWork, Record, Comment, Status, LibraryEntry].each do |resource_class|
        update_or_delete_resource(resource_class, target_work, original_work)
      end

      target_work.destroy_in_batches
    end
  end

  def update_or_delete_pol_resource(resource_class, column, target_work, original_work)
    resource_class.where(column.to_sym => target_work).find_each do |t_resource|
      o_resource = resource_class.where(:user => t_resource.user, column.to_sym => original_work).first

      if o_resource.blank?
        t_resource.update_column("#{column}_id".to_sym, original_work.id)
      else
        t_resource.destroy
      end
    end
  end

  def update_or_delete_resource(resource_class, target_work, original_work)
    resource_class.where(work: target_work).find_each do |t_resource|
      o_resource = resource_class.where(user: t_resource.user, work: original_work).first

      if o_resource.blank?
        t_resource.update_column(:work_id, original_work.id)
      else
        t_resource.destroy
      end
    end
  end

  task update_score: :environment do
    RATE_MAX = 100

    Anime.only_kept.find_each do |w|
      episodes = w.episodes.only_kept.where.not(satisfaction_rate: nil)
      ratings_count = episodes.pluck(:ratings_count).inject(&:+)
      rates = episodes.pluck(:satisfaction_rate)

      if rates.all?(&:nil?)
        w.update_column(:satisfaction_rate, nil)
        next
      end

      rates_count = rates.length
      rates_sum = rates.inject(&:+)
      rates_avg = rates_sum.to_f / rates_count
      satisfaction_rate = (rates_avg / RATE_MAX * 100).round(2)

      outputs = [
        "rates_count: #{rates_count}",
        "rates_sum: #{rates_sum}",
        "rates_avg: #{rates_avg}",
        "satisfaction_rate: #{satisfaction_rate}"
      ]
      puts "Work: #{w.id} => #{outputs.join(", ")}"

      w.update_columns(satisfaction_rate: satisfaction_rate, ratings_count: ratings_count)
    end
  end
end
