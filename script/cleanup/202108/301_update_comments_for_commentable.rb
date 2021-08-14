# frozen_string_literal: true

Comment.where(commentable_id: nil, commentable_type: nil).preload(:episode_record).each do |c|
  p "comments.id: #{c.id}"

  c.update_columns(commentable_id: c.episode_record.record_id, commentable_type: "Record")
end
