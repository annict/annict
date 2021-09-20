# frozen_string_literal: true

WorkTaggable.preload(:user, :work_tag).find_each do |work_taggable|
  ActiveRecord::Base.transaction do
    puts "work_taggable: #{work_taggable.id}"

    user = work_taggable.user
    work_tag = work_taggable.work_tag
    description = work_taggable.description.presence || ""

    collection = user.collections.where(name: work_tag.name).first_or_create!(
      created_at: work_taggable.created_at,
      updated_at: work_taggable.updated_at
    )
    if description != collection.description
      collection.update_columns(description: description)
    end

    user.work_taggings.where(work_tag: work_tag).preload(:work).each do |work_tagging|
      work = work_tagging.work

      user.collection_items.where(collection: collection, work: work).first_or_create!(
        name: work.title,
        created_at: work_tagging.created_at,
        updated_at: work_tagging.updated_at
      )
    end
  end
end

Collection.find_each do |collection|
  puts "collection: #{collection.id}"
  collection.update_columns(collection_items_count: collection.collection_items.count)
end
