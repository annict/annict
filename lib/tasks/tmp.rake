# frozen_string_literal: true

namespace :tmp do
  task move_from_collection_to_tag: :environment do
    ActiveRecord::Base.transaction do
      Collection.find_each do |c|
        puts "collection: #{c.id}"

        work_tag = WorkTag.where(name: c.title).first_or_create!

        c.collection_items.each do |ci|
          c.user.work_taggings.where(work: ci.work, work_tag: work_tag).first_or_create!
        end
      end
    end
  end

  task move_from_collection_item_to_work_comment: :environment do
    ActiveRecord::Base.transaction do
      CollectionItem.find_each do |ci|
        puts "collection item: #{ci.id}"

        work_comment = WorkComment.where(user: ci.user, work: ci.work).first_or_initialize

        if work_comment.body.present? && !work_comment.body.include?(ci.comment)
          work_comment.body += "\n#{ci.comment}"
        else
          work_comment.body = ci.comment
        end

        work_comment.save! if work_comment.body.present?
      end
    end
  end
end
