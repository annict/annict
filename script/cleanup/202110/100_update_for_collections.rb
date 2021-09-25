# frozen_string_literal: true

WorkTaggable.preload(:user, :work_tag).find_in_batches do |work_taggables|
  attributes = work_taggables.map do |work_taggable|
    user = work_taggable.user
    work_tag = work_taggable.work_tag
    description = work_taggable.description.presence || ""

    {
      description: description,
      name: work_tag.name,
      created_at: work_taggable.created_at,
      updated_at: work_taggable.updated_at,
      user_id: user.id
    }
  end

  result = Collection.insert_all!(attributes)
  puts "Upserted: #{result.rows.first(3)}..."
end
