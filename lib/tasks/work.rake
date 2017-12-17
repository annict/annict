# frozen_string_literal: true

namespace :work do
  task notify_untouched_works: :environment do
    works = Work.where(auto_episodes_count: 0).order(watchers_count: :desc).limit(3)
    WorkMailer.untouched_works_notification(works.pluck(:id)).deliver_later
  end

  # 指定したWorkを削除する
  # コマンド実行例: rake work:hide_overlapped_work[4458,4485]
  task :hide_overlapped_work, %i(target_work_id original_work_id) => :environment do |_, args|
    # 削除対象のWork
    target_work = Work.find(args[:target_work_id])
    # オリジナルのWork
    original_work = Work.find(args[:original_work_id])

    ActiveRecord::Base.transaction do
      [
        { resource_class: Activity, column: :recipient }
      ].each do |hash|
        update_or_delete_pol_resource(hash[:resource_class], hash[:column], target_work, original_work)
      end

      [ChannelWork, Checkin, Comment, Status, LatestStatus].each do |resource_class|
        update_or_delete_resource(resource_class, target_work, original_work)
      end

      target_work.hide!
    end
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

  task :send_next_season_came_email, %i(season_year season_name all) => :environment do |t, args|
    Rails.logger = Logger.new(STDOUT) if Rails.env.development?
    Rails.logger.info "work:send_next_season_came_email >> task started"

    users = if args[:all] == "true"
      User.
        joins(:email_notification).
        where(email_notifications: { event_next_season_came: true })
    else
      User.where(username: "shimbaco")
    end

    users.find_each do |user|
      Rails.logger.info "work:send_next_season_came_email >> user: #{user.id}"
      EmailNotificationService.send_email(
        "next_season_came", user, args[:season_year], args[:season_name]
      )
    end

    Rails.logger.info "work:send_next_season_came_email >> task processed"
  end

  task send_favorite_works_added_email: :environment do
    cast_work_ids = Cast.published.yesterday.pluck(:work_id)
    staff_work_ids = Staff.published.yesterday.pluck(:work_id)
    season = Season.find_by_slug(ENV.fetch("ANNICT_CURRENT_SEASON"))
    works = Work.published.where(id: (cast_work_ids | staff_work_ids))
    works = works.
      where("season_year >= ? AND season_name > ?", season.year, season.name_value).
      or(works.where("season_year > ?", season.year)).
      or(works.where(season_year: season.year, season_name: nil)).
      or(works.where(season_year: nil))

    works.find_each do |work|
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

  task update_score: :environment do
    results = {}
    episode_score_max = 10

    Work.published.find_each do |w|
      scores = w.episodes.published.pluck(:score)

      if scores.all?(&:nil?)
        w.update_column(:score, nil)
        next
      end

      scores = scores.map { |s| s.nil? ? 0 : s + 1 }
      score_avg = (scores.inject(&:+) / scores.length) + 1
      score_range = 1..(episode_score_max + 2)
      wilson_score = WilsonScore.rating_lower_bound(score_avg, w.watchers_count, score_range)

      puts "Work: #{w.id} => #{wilson_score}"

      results[w.id] = wilson_score
    end

    wilson_score_max = results.values.max

    results.each do |work_id, wilson_score|
      score = (wilson_score / wilson_score_max * 10).round(2)

      puts "Work: #{work_id} => score: #{score}"

      Work.update(work_id, score: score)
    end
  end
end
