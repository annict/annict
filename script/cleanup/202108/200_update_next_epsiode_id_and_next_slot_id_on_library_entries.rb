# frozen_string_literal: true

users = User.only_kept.past_week(field: :current_sign_in_at)
# users = users.where(username: "shimbaco")

users.find_each do |user|
  p "user: #{user.id}"

  user.library_entries.preload(:work, :program).watching.find_each do |le|
    work = le.work
    program = le.program
    next_episode = work.episodes.only_kept.where.not(id: le.watched_episode_ids).order(:sort_number).first
    next_slot = program&.slots&.only_kept&.find_by(episode: next_episode)

    le.next_episode = next_episode
    le.next_slot = next_slot

    le.save!
  end
end
