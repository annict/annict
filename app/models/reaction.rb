# frozen_string_literal: true

# == Schema Information
#
# Table name: reactions
#
#  id                 :bigint           not null, primary key
#  kind               :string           not null
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#  collection_item_id :bigint
#  target_user_id     :bigint           not null
#  user_id            :bigint           not null
#
# Indexes
#
#  index_reactions_on_collection_item_id  (collection_item_id)
#  index_reactions_on_target_user_id      (target_user_id)
#  index_reactions_on_user_id             (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (collection_item_id => collection_items.id)
#  fk_rails_...  (target_user_id => users.id)
#  fk_rails_...  (user_id => users.id)
#

class Reaction < ApplicationRecord
  extend Enumerize

  enumerize :kind, in: %w[
    thumbs_up
  ]

  validates :kind, presence: true

  belongs_to :user
  belongs_to :target_user, class_name: "User"
end
