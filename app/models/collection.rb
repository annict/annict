# frozen_string_literal: true

# == Schema Information
#
# Table name: collections
#
#  id                     :bigint           not null, primary key
#  collection_items_count :integer          default(0), not null
#  deleted_at             :datetime
#  description            :string           default(""), not null
#  likes_count            :integer          default(0), not null
#  name                   :string           not null
#  created_at             :datetime         not null
#  updated_at             :datetime         not null
#  user_id                :bigint           not null
#
# Indexes
#
#  index_collections_on_deleted_at        (deleted_at)
#  index_collections_on_user_id           (user_id)
#  index_collections_on_user_id_and_name  (user_id,name) UNIQUE
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#

class Collection < ApplicationRecord
  include SoftDeletable

  belongs_to :user
  has_many :collection_items, dependent: :destroy
  has_many :works, through: :collection_items

  validates :name, presence: true, length: {maximum: 50}
  validates :description, length: {maximum: 500}

  def contain?(work)
    collection_items.where(work: work).exists?
  end

  def positions_for_select
    collection_items.only_kept.order(:position).map do |item|
      key = item.position.to_s
      key += " (#{I18n.t("messages.collections.position_of_x", item_title: item.title)})"
      [key, item.position]
    end
  end
end
