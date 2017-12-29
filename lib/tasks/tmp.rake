# frozen_string_literal: true

namespace :tmp do
  task set_locale: :environment do
    [
      { model: Checkin, column: :comment },
      { model: Comment, column: :body },
      { model: DbComment, column: :body },
      { model: ForumComment, column: :body },
      { model: ForumPost, column: :body },
      { model: Item, column: :title },
      { model: Review, column: :body },
      { model: UserlandProject, column: :description },
      { model: WorkComment, column: :body },
      { model: WorkTag, column: :name },
      { model: WorkTaggable, column: :description }
    ].each do |h|
      h[:model].where.not(h[:column] => [nil, ""]).find_each do |record|
        puts "#{record.class.name}: #{record.id}"
        record.detect_locale!(h[:column])
        record.save!
      end
    end
  end

  task set_allowed_locales: :environment do
    User.update_all(allowed_locales: ApplicationRecord::LOCALES)
  end

  task update_tips: :environment do
    tip1 = Tip.find_by(slug: "status")
    tip1.update_columns(locale: :ja, icon_name: "fa-info-circle")

    tip2 = Tip.find_by(slug: "channel")
    tip2.update_columns(locale: :ja, icon_name: "fa-info-circle")

    tip3 = Tip.find_by(slug: "checkin")
    tip3.update_columns(slug: "record", locale: :ja, icon_name: "fa-info-circle")

    Tip.create(target: 0, slug: "status", title: "Change Statuses of Works", icon_name: "fa-info-circle", locale: :en)
    Tip.create(target: 0, slug: "record", title: "Track What Episodes You Watched", icon_name: "fa-info-circle", locale: :en)
  end
end
