# frozen_string_literal: true

namespace :email_notification do
  task notify_untouched_works: :environment do
    works = Work.where(auto_episodes_count: 0).order(watchers_count: :desc).limit(3)
    WorkMailer.untouched_works_notification(works.pluck(:id)).deliver_later
  end

  task send_favorite_works_added_email: :environment do
    casts = Cast.only_kept.past_week
    staffs = Staff.only_kept.past_week

    works = Work.only_kept.where(id: casts.pluck(:work_id) | staffs.pluck(:work_id)).gt_current_season

    next if works.blank?

    users = User.
      only_kept.
      joins(:email_notification).
      where(email_notifications: { event_favorite_works_added: true })

    users.find_each do |user|
      favorite_character_ids = user.favorite_characters.only_kept.pluck(:id)
      favorite_person_ids = user.favorite_people.only_kept.pluck(:id)
      favorite_organization_ids = user.favorite_organizations.only_kept.pluck(:id)

      next if favorite_character_ids.blank? && favorite_person_ids.blank? && favorite_organization_ids.blank?

      character_works = works.joins(:casts).where(casts: { character_id: favorite_character_ids })
      cast_person_works = works.joins(:cast_people).where(casts: { person_id: favorite_person_ids })
      staff_person_works = works.joins(:staff_people).where(staffs: { resource_id: favorite_person_ids })
      organization_works = works.joins(:organizations).where(staffs: { resource_id: favorite_organization_ids })

      work_ids = (character_works | cast_person_works | staff_person_works | organization_works).pluck(:id)
      library_work_ids = user.statuses.where(work_id: work_ids).pluck(:work_id)
      target_work_ids = work_ids - library_work_ids

      next if target_work_ids.blank?

      EmailNotificationService.send_email("favorite_works_added", user, target_work_ids)
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
