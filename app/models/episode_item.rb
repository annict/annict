# frozen_string_literal: true
# == Schema Information
#
# Table name: episode_items
#
#  id         :integer          not null, primary key
#  episode_id :integer          not null
#  item_id    :integer          not null
#  user_id    :integer          not null
#  work_id    :integer          not null
#  aasm_state :string           default("published"), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_episode_items_on_episode_id              (episode_id)
#  index_episode_items_on_episode_id_and_item_id  (episode_id,item_id) UNIQUE
#  index_episode_items_on_item_id                 (item_id)
#  index_episode_items_on_user_id                 (user_id)
#  index_episode_items_on_work_id                 (work_id)
#

class EpisodeItem < ApplicationRecord
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :episode
  belongs_to :item
  belongs_to :user
  belongs_to :work
end
