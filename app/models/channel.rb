# frozen_string_literal: true

# == Schema Information
#
# Table name: channels
#
#  id               :integer          not null, primary key
#  channel_group_id :integer          not null
#  sc_chid          :integer          not null
#  name             :string           not null
#  published        :boolean          default(TRUE), not null
#  created_at       :datetime
#  updated_at       :datetime
#
# Indexes
#
#  channels_channel_group_id_idx  (channel_group_id)
#  channels_sc_chid_key           (sc_chid) UNIQUE
#

class Channel < ApplicationRecord
  belongs_to :channel_group
  has_many :programs

  scope(:published, -> { where(published: true) })

  def self.fastest(work)
    receivable_channel_ids = pluck(:id)

    if receivable_channel_ids.present? && work.episodes.present?
      conditions = { channel_id: receivable_channel_ids, episode: work.episodes.first }
      fastest_program = Program.where(conditions).order(:started_at).first

      fastest_program.present? ? fastest_program.channel : nil
    end
  end
end
