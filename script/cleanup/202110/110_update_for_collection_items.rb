# frozen_string_literal: true

WorkTagging.preload(:user, :work, :work_tag).find_in_batches(batch_size: 3_000) do |work_taggings|
  attributes = work_taggings.map do |work_tagging|
    user = work_tagging.user
    work = work_tagging.work
    work_tag = work_tagging.work_tag
    collection = user.collections.find_by!(name: work_tag.name)

    {
      name: work.title,
      created_at: work_tagging.created_at,
      updated_at: work_tagging.updated_at,
      collection_id: collection.id,
      user_id: user.id,
      work_id: work.id
    }
  end

  result = CollectionItem.insert_all!(attributes)
  puts "Upserted: #{result.rows.first(3)}..."
end
