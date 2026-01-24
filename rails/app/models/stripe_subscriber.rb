# typed: false
# frozen_string_literal: true

class StripeSubscriber < ApplicationRecord
  ACTIVE_STATUSES = %w[active past_due].freeze

  validates :stripe_customer_id, presence: true
  validates :stripe_subscription_id, presence: true
  validates :stripe_price_id, presence: true
  validates :stripe_status, presence: true
  validates :stripe_current_period_start, presence: true
  validates :stripe_current_period_end, presence: true

  # active または past_due 状態をアクティブとして扱う
  # past_due は支払い遅延中だが、Stripeがリトライ中のため猶予期間として利用可能
  def active?
    stripe_status.in?(ACTIVE_STATUSES)
  end
end
