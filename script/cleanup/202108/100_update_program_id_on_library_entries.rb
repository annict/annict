# frozen_string_literal: true

users = User.only_kept.past_year(field: :current_sign_in_at)
# users = users.where(username: "shimbaco")

users.find_each do |user|
  p "user: #{user.id}"

  ActiveRecord::Base.transaction do
    ChannelAnime.where(user: user).find_each do |ca|
      program = Program.where(work_id: ca.work_id, channel_id: ca.channel_id).order(started_at: :desc).first
      next unless program

      library_entry = user.library_entries.where(work_id: ca.work_id).first
      next unless library_entry
      next if library_entry.program_id

      library_entry.update!(program_id: program.id)
    end
  end
end
