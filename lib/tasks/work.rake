namespace :work do
  task notify_untouched_works: :environment do
    works = Work.where(episodes_count: 0).order(watchers_count: :desc).limit(3)
    WorkMailer.delay.untouched_works_notification(works.pluck(:id))
  end

  # 指定したWorkを削除する
  # コマンド実行例: rake work:delete_overlapped_work[4458,4485]
  task :delete_overlapped_work, [:target_work_id, :original_work_id] => :environment do |t, args|
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

  task :send_next_season_came_email, %i(season_year season_name) => :environment do |t, args|
    Rails.logger = Logger.new(STDOUT) if Rails.env.development?
    Rails.logger.info "work:send_next_season_came_email >> task started"

    season = Season.where(year: args[:season_year], name: args[:season_name]).first
    if season.blank?
      Rails.logger.info "work:send_next_season_came_email >> no season found"
      next
    end
    Rails.logger.info "work:send_next_season_came_email >> season: #{season.slug}"

    users = User.
      joins(:email_notification).
      where(email_notifications: { event_next_season_came: true })

    users.find_each do |user|
      Rails.logger.info "work:send_next_season_came_email >> user: #{user.id}"
      EmailNotificationService.send_email("next_season_came", user, season.id)
    end

    Rails.logger.info "work:send_next_season_came_email >> task processed"
  end

  task send_favorite_works_added_email: :environment do
    Work.yesterday.find_each do |work|
      favorite_character_user_ids = FavoriteCharacter.
        joins(:character).
        merge(work.characters).
        pluck(:user_id)
      favorite_people_user_ids = FavoritePerson.
        joins(:person).
        merge(work.people).
        pluck(:user_id)
      favorite_org_user_ids = FavoriteOrganization.
        joins(:organization).
        merge(work.organizations).
        pluck(:user_id)
      user_ids = favorite_character_user_ids |
        favorite_people_user_ids |
        favorite_org_user_ids
      users = User.
        joins(:email_notification).
        where(id: user_ids).
        where(email_notifications: { event_favorite_works_added: true })

      users.find_each do |user|
        next if user.statuses.where(work: work).exists?
        EmailNotificationService.send_email("favorite_works_added", user, work.id)
      end
    end
  end
end
