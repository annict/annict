# frozen_string_literal: true

namespace :tmp do
  task move_single_works_records_to_reviews: :environment do
    ActiveRecord::Base.transaction do
      Work.published.find_each do |w|
        next unless w.single?

        puts "--- work: #{w.id}"

        w.update_column(:no_episodes, true)

        episode = w.episodes.first

        episode.update_column(:aasm_state, "hidden")

        episode.records.each do |r|
          r.update_column(:aasm_state, "hidden")

          next if r.comment.blank?

          puts "- record: #{r.id}"

          review = Review.create(
            user_id: r.user_id,
            work_id: r.work_id,
            body: r.comment,
            rating_overall_state: r.rating_state,
            created_at: r.created_at,
            updated_at: r.updated_at
          )

          r.update_attributes(review: review)

          r.activities.each do |a|
            a.update_attributes(recipient: w, trackable: review, action: "create_review")
          end

          next if r.likes.blank?

          r.likes.each do |l|
            puts "- like: #{l.id}"

            l.update_attributes(recipient: review)
          end
        end
      end
    end
  end
end
