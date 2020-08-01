# frozen_string_literal: true
# == Schema Information
#
# Table name: collection_items
#
#  id              :bigint           not null, primary key
#  aasm_state      :string           default("published"), not null
#  comment         :text
#  deleted_at      :datetime
#  position        :integer          default(0), not null
#  reactions_count :integer          default(0), not null
#  title           :string           not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#  anime_id        :bigint           not null
#  collection_id   :bigint           not null
#  user_id         :bigint           not null
#
# Indexes
#
#  index_collection_items_on_anime_id                    (anime_id)
#  index_collection_items_on_collection_id               (collection_id)
#  index_collection_items_on_collection_id_and_anime_id  (collection_id,anime_id) UNIQUE
#  index_collection_items_on_deleted_at                  (deleted_at)
#  index_collection_items_on_user_id                     (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (anime_id => animes.id)
#  fk_rails_...  (collection_id => collections.id)
#  fk_rails_...  (user_id => users.id)
#

class CollectionItem < ApplicationRecord
  include SoftDeletable

  acts_as_list scope: :collection_id

  belongs_to :user
  belongs_to :collection, touch: true
  belongs_to :work
  has_many :reactions, dependent: :destroy

  validates :title, presence: true, length: { maximum: 50 }
  validates :comment, length: { maximum: 1000 }
end
