# typed: false
# frozen_string_literal: true

class GumroadSubscriber < ApplicationRecord
  validates :gumroad_id, presence: true
  validates :gumroad_product_id, presence: true

  def active?
    !gumroad_cancelled_at&.past? && !gumroad_ended_at&.past?
  end
end
