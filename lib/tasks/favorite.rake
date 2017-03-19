# frozen_string_literal: true

namespace :favorite do
  task set_watched_counter: :environment do
    User.find_each do |u|
      puts "user: #{u.id}"

      statuses = u.statuses.work_published.with_kind(:watched)

      u.favorite_people.each do |favorite_person|
        cast_work_ids = favorite_person.person.cast_works.pluck(:id)
        staff_work_ids = favorite_person.person.staff_works.pluck(:id)
        count = statuses.where(work_id: (cast_work_ids | staff_work_ids)).count
        favorite_person.update_column(:watched_works_count, count)
      end

      u.favorite_organizations.each do |favorite_org|
        staff_work_ids = favorite_org.organization.staff_works.pluck(:id)
        count = statuses.where(work_id: staff_work_ids).count
        favorite_org.update_column(:watched_works_count, count)
      end
    end
  end
end
