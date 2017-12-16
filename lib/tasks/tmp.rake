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
end
