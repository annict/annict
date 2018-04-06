# frozen_string_literal: true

namespace :tmp do
  task merge_reviews_with_records: :environment do
    Review.find_each do |review|
      ActiveRecord::Base.transaction do
        puts review.id

        puts "create record..."
        attr = {
            user: review.user,
            work: review.work,
            comment: review.body,
            rating_state: review.rating_overall_state,
            likes_count: review.likes_count,
            impressions_count: review.impressions_count,
            review: review,
            aasm_state: review.aasm_state,
            modify_comment: review.modified_at.present?,
            created_at: review.created_at,
            updated_at: review.updated_at,
            oauth_application_id: review.oauth_application_id,
            locale: review.locale
        }
        record = Record.where(user: review.user, review: review).published.first_or_create!(attr)

        puts "update review..."
        review.update_column(:record_id, record.id)

        puts "update activities..."
        review.activities.update_all(trackable_type: "Record", trackable_id: record.id, action: :create_record, record_id: record.id)

        puts "update likes..."
        review.likes.update_all(recipient_type: "Record", recipient_id: record.id)
      end
    end
  end
end
