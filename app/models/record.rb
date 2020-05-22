# frozen_string_literal: true
# == Schema Information
#
# Table name: records
#
#  id                :bigint           not null, primary key
#  aasm_state        :string           default("published"), not null
#  deleted_at        :datetime
#  impressions_count :integer          default(0), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#  user_id           :bigint           not null
#  work_id           :bigint           not null
#
# Indexes
#
#  index_records_on_deleted_at  (deleted_at)
#  index_records_on_user_id     (user_id)
#  index_records_on_work_id     (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#

class Record < ApplicationRecord
  include SoftDeletable

  RATING_STATES = %i(bad average good great).freeze

  counter_culture :user
  counter_culture :work

  belongs_to :user
  belongs_to :work
  has_one :episode_record, dependent: :destroy
  has_one :work_record, dependent: :destroy

  def episode_record?
    episode_record.present?
  end
end
