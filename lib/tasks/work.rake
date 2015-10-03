namespace :work do
  task notify_untouched_works: :environment do
    works = Work.where(episodes_count: 0).order(watchers_count: :desc).limit(3)
    WorkMailer.delay.untouched_works_notification(works.pluck(:id))
  end

  # 指定したWorkを削除する
  # コマンド実行例: rake work:delete_overlapped_work[4458,4485]
  task :delete_overlapped_work, [:target_work_id, :original_work_id] do |t, args|
    # 削除対象のWork
    target_work = Work.find(args[:target_work_id])
    # オリジナルのWork
    original_work = Work.find(args[:original_work_id])

    [{ resource_class: Activity, column: :recipient }].each do |hash|
      update_or_delete_pol_resource(hash[:resource_class], hash[:column], target_work, original_work)
    end

    [ChannelWork, Checkin, Check, Comment, Status].each do |resource_class|
      update_or_delete_resource(resource_class, target_work, original_work)
    end

    target_work.destroy
  end

  def update_or_delete_pol_resource(resource_class, column, target_work, original_work)
    resource_class.where(column.to_sym => target_work).find_each do |t_resource|
      o_resource = resource_class.where(user: t_resource.user, column.to_sym => original_work).first

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
end
