# frozen_string_literal: true
# == Schema Information
#
# Table name: channels
#
#  id                :integer          not null, primary key
#  channel_group_id  :integer          not null
#  sc_chid           :integer
#  name              :string           not null
#  created_at        :datetime
#  updated_at        :datetime
#  streaming_service :boolean          default(FALSE)
#  aasm_state        :string           default("published"), not null
#
# Indexes
#
#  channels_channel_group_id_idx        (channel_group_id)
#  channels_sc_chid_key                 (sc_chid) UNIQUE
#  index_channels_on_streaming_service  (streaming_service)
#

class Channel < ApplicationRecord
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      after do
        program_details.published.each(&:hide!)
      end

      transitions from: :published, to: :hidden
    end
  end

  belongs_to :channel_group
  has_many :program_details, dependent: :destroy
  has_many :programs, dependent: :destroy

  scope :with_streaming_service, -> { where(streaming_service: true) }

  def self.fastest(work)
    receivable_channel_ids = pluck(:id)

    if receivable_channel_ids.present? && work.episodes.present?
      conditions = { channel_id: receivable_channel_ids, episode: work.episodes.first }
      fastest_program = Program.where(conditions).order(:started_at).first

      fastest_program.present? ? fastest_program.channel : nil
    end
  end
end
