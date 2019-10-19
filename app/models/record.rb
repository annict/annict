# frozen_string_literal: true
# == Schema Information
#
# Table name: records
#
#  id                :bigint           not null, primary key
#  aasm_state        :string           default("published"), not null
#  impressions_count :integer          default(0), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#  user_id           :bigint           not null
#  work_id           :bigint           not null
#
# Indexes
#
#  index_records_on_user_id  (user_id)
#  index_records_on_work_id  (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#

class Record < ApplicationRecord
  include AASM

  RATING_STATES = %i(bad average good great).freeze

  is_impressionable counter_cache: true, unique: true

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :user, counter_cache: true
  belongs_to :work, counter_cache: true
  has_one :episode_record, dependent: :destroy
  has_one :work_record, dependent: :destroy

  def episode_record?
    episode_record.present?
  end
end
