# frozen_string_literal: true

namespace :email_notification do
  task notify_untouched_works: :environment do
    works = Work.where(auto_episodes_count: 0).order(watchers_count: :desc).limit(3)
    WorkMailer.untouched_works_notification(works.pluck(:id)).deliver_later
  end

  task send_favorite_works_added_email: :environment do
    cast_work_ids = Cast.only_kept.yesterday.pluck(:work_id)
    staff_work_ids = Staff.only_kept.yesterday.pluck(:work_id)
    works = Work.only_kept.where(id: (cast_work_ids | staff_work_ids)).gt_current_season

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
        only_kept.
        joins(:email_notification).
        where(id: user_ids).
        where(email_notifications: { event_favorite_works_added: true })

      users.find_each do |user|
        next if user.statuses.where(work: work).exists?

        EmailNotificationService.send_email("favorite_works_added", user, work.id)
      end
    end
  end

  task send_related_works_added_email: :environment do
    works = Work.only_kept.yesterday.gt_current_season

    next if works.blank?

    users = User.
      only_kept.
      joins(:email_notification).
      where(email_notifications: { event_related_works_added: true })

    works.find_each do |work|
      series_ids = work.series_list.pluck(:id)
      related_work_ids = SeriesWork.where(series_id: series_ids).pluck(:work_id).uniq - [work.id]

      next if related_work_ids.blank?

      users.find_each do |user|
        positive_statuses = user.library_entries.positive

        next unless positive_statuses.exists?
        next if user.statuses.pluck(:work_id).include?(work.id)

        if (positive_statuses.pluck(:work_id) & related_work_ids).present?
          EmailNotificationService.send_email("related_works_added", user, work.id)
        end
      end
    end
  end
end
