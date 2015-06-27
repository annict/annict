# == Schema Information
#
# Table name: draft_episodes
#
#  id              :integer          not null, primary key
#  episode_id      :integer          not null
#  work_id         :integer          not null
#  number          :string
#  sort_number     :integer          default(0), not null
#  title           :string
#  next_episode_id :integer
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#  prev_episode_id :integer
#
# Indexes
#
#  index_draft_episodes_on_episode_id       (episode_id)
#  index_draft_episodes_on_next_episode_id  (next_episode_id)
#  index_draft_episodes_on_prev_episode_id  (prev_episode_id)
#  index_draft_episodes_on_work_id          (work_id)
#

class DraftEpisode < ActiveRecord::Base
  include DraftCommon
  include EpisodeCommon

  belongs_to :origin, class_name: "Episode", foreign_key: :episode_id
  belongs_to :work

  validates :title, presence: true
end
