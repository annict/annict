# frozen_string_literal: true

users = User.only_kept.past_year(field: :current_sign_in_at)
# users = users.where(username: "shimbaco")

users.find_each do |user|
  p "user: #{user.id}"

  ActiveRecord::Base.transaction do
    user.channel_works.find_each do |cw|
      program = Program.find_by(channel_id: cw.channel_id, work_id: cw.work_id)

      next unless program

      library_entry = user.library_entries.where(work_id: cw.work_id).first_or_create!
      library_entry.update!(program_id: program.id)
    end
  end
end
