# frozen_string_literal: true

users = User.only_kept.past_year(field: :current_sign_in_at)
# users = users.where(username: "shimbaco")

users.find_each do |user|
  p "user: #{user.id}"

  ActiveRecord::Base.transaction do
    user.channel_works.find_each do |cw|
      program = Program.find_by(channel_id: cw.channel_id, work_id: cw.work_id)

      next unless program

      user.user_programs.where(work_id: cw.work_id).first_or_create!(program_id: program.id, created_at: cw.created_at, updated_at: cw.updated_at)
    end
  end
end
