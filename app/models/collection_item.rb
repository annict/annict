# frozen_string_literal: true

# == Schema Information
#
# Table name: collection_items
#
#  id            :bigint           not null, primary key
#  body          :text             default(""), not null
#  deleted_at    :datetime
#  likes_count   :integer          default(0), not null
#  name          :string           not null
#  position      :integer          default(0), not null
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#  collection_id :bigint           not null
#  user_id       :bigint           not null
#  work_id       :bigint           not null
#
# Indexes
#
#  index_collection_items_on_collection_id              (collection_id)
#  index_collection_items_on_collection_id_and_work_id  (collection_id,work_id) UNIQUE
#  index_collection_items_on_deleted_at                 (deleted_at)
#  index_collection_items_on_user_id                    (user_id)
#  index_collection_items_on_work_id                    (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (collection_id => collections.id)
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#

class CollectionItem < ApplicationRecord
  include SoftDeletable

  acts_as_list scope: :collection_id

  belongs_to :user
  belongs_to :collection
  belongs_to :work
  has_many :likes, dependent: :destroy

  validates :body, length: {maximum: 1000}
end
