# frozen_string_literal: true
# == Schema Information
#
# Table name: reactions
#
#  id                 :integer          not null, primary key
#  user_id            :integer          not null
#  target_user_id     :integer          not null
#  kind               :string           not null
#  collection_item_id :integer
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#
# Indexes
#
#  index_reactions_on_collection_item_id  (collection_item_id)
#  index_reactions_on_target_user_id      (target_user_id)
#  index_reactions_on_user_id             (user_id)
#

class Reaction < ApplicationRecord
  extend Enumerize

  enumerize :kind, in: %w(
    thumbs_up
  )

  validates :kind, presence: true

  belongs_to :user
  belongs_to :target_user, class_name: "User"
end
