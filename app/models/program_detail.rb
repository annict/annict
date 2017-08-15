# frozen_string_literal: true
# == Schema Information
#
# Table name: program_details
#
#  id                :integer          not null, primary key
#  channel_id        :integer          not null
#  work_id           :integer          not null
#  latest_program_id :integer
#  url               :string
#  started_at        :datetime
#  repeat_on         :string
#  aasm_state        :string           default("published"), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#
# Indexes
#
#  index_program_details_on_channel_id              (channel_id)
#  index_program_details_on_channel_id_and_work_id  (channel_id,work_id) UNIQUE
#  index_program_details_on_latest_program_id       (latest_program_id)
#  index_program_details_on_work_id                 (work_id)
#

class ProgramDetail < ApplicationRecord
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(channel_id work_id url started_at).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  validates :url, url: { allow_blank: true }

  belongs_to :channel
  belongs_to :work
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
end
