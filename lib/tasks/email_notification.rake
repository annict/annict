# frozen_string_literal: true

namespace :email_notification do
  task notify_untouched_works: :environment do
    works = Work.where(auto_episodes_count: 0).order(watchers_count: :desc).limit(3)
    WorkMailer.untouched_works_notification(works.pluck(:id)).deliver_later
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
end