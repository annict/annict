# frozen_string_literal: true

WorkComment.preload(:user, :work).find_each do |work_comment|
  ActiveRecord::Base.transaction do
    puts "work_comment: #{work_comment.id}"

    user = work_comment.user
    work = work_comment.work

    library_entry = user.library_entries.where(work: work).first_or_create!
    library_entry.update_columns(note: work_comment.body)
  end
end
