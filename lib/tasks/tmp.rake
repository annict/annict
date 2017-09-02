# frozen_string_literal: true

namespace :tmp do
  task append_review_title_to_body: :environment do
    ActiveRecord::Base.transaction do
      Review.where.not(title: "").find_each do |r|
        puts r.id
        r.body = "#{r.title}\n\n#{r.body}"
        r.save!
      end
    end
  end
end
