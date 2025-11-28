# typed: false
# frozen_string_literal: true

class Collection < ApplicationRecord
  include SoftDeletable

  belongs_to :user
  has_many :collection_items, dependent: :destroy
  has_many :works, through: :collection_items

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
